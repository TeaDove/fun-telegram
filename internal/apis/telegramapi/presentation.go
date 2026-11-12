package telegramapi

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/glebarez/sqlite"
	"github.com/teadove/teasutils/utils/redactutils"

	"github.com/TeaDove/fun-telegram/internal/services/analyticsservice"
	"github.com/TeaDove/fun-telegram/internal/utils"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	protoerrors "github.com/celestix/gotgproto/errors"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/cockroachdb/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Presentation struct {
	telegramAPI *tg.Client
	protoClient *gotgproto.Client

	router map[string]messageProcessor

	analiticsService *analyticsservice.Service
}

func NewProtoClient(ctx context.Context, settings utils.TelegramSettings) (*gotgproto.Client, error) {
	middlewares := make([]telegram.Middleware, 0, 2)

	if settings.RateLimiterEnabled {
		middlewares = append(middlewares, ratelimit.New(
			rate.Every(settings.RateLimiterRate),
			settings.RateLimiterLimit))
	}

	var runMiddleware func(
		origRun func(ctx context.Context, f func(ctx context.Context) error) (err error),
		ctx context.Context,
		f func(ctx context.Context) (err error),
	) (err error)

	if settings.FloodWaiterEnabled {
		waiter := floodwait.
			NewSimpleWaiter().
			WithMaxWait(time.Minute * 10).
			WithMaxRetries(20)

		middlewares = append(middlewares, waiter)
	}

	zerolog.Ctx(ctx).Info().
		Str("app_id", redactutils.Redact(strconv.Itoa(settings.AppID))).
		Str("app_hash", redactutils.Redact(settings.AppHash)).
		Msg("creating.client")

	options := &gotgproto.ClientOpts{
		Context:          ctx,
		DisableCopyright: true,
		Session:          sessionMaker.SqlSession(sqlite.Open(sessionPath)),
		NoAutoAuth:       true,
		Middlewares:      middlewares,
		RunMiddleware:    runMiddleware,
		RetryInterval:    10 * time.Second,
		MaxRetries:       10,
		DC:               2,
	}
	clientType := gotgproto.ClientTypePhone("")
	client, err := gotgproto.NewClient(settings.AppID, settings.AppHash, clientType, options)

	if errors.Is(err, protoerrors.ErrSessionUnauthorized) {
		if client != nil {
			client.Stop()
		}

		err = LoginQR(ctx, settings, os.Stdin, os.Stdout)
		if err != nil {
			return nil, errors.Wrap(err, "authorize Telegram session")
		}

		client, err = gotgproto.NewClient(settings.AppID, settings.AppHash, clientType, options)
	}

	if err != nil {
		if client != nil {
			client.Stop()
		}

		return nil, errors.Wrap(err, "create gotgproto client")
	}

	return client, nil
}

func MustNewTelegramPresentation(
	protoClient *gotgproto.Client,
	analiticsService *analyticsservice.Service,
) *Presentation {
	api := protoClient.API()

	presentation := Presentation{
		protoClient:      protoClient,
		telegramAPI:      api,
		analiticsService: analiticsService,
	}

	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback: presentation.injectContext,
			Outgoing: true,
		},
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.route,
			Filters:       filters.Message.Text,
			UpdateFilters: filterNonNewMessages,
			Outgoing:      true,
		},
	)

	presentation.router = map[string]messageProcessor{
		"help": {
			executor:    presentation.helpCommandHandler,
			description: "get this message",
			flags:       []optFlag{},
		},
		"stats": {
			executor:    presentation.statsCommand,
			description: "render charts from the selected chat",
			flags:       historyFlags(),
			example:     "-d=7 -q",
		},
		"dump-history": {
			executor:    presentation.dumpHistoryCommand,
			description: "send compressed chat history to Saved Messages",
			flags:       historyFlags(),
			example:     `"Chat Title" -d=7`,
		},
		"restart": {
			executor:    presentation.restartCommandHandler,
			description: "restarts bot",
		},
	}

	dp, ok := protoClient.Dispatcher.(*dispatcher.NativeDispatcher)
	if !ok {
		panic("telegram dispatcher is not native")
	}

	dp.Error = presentation.errorHandler
	dp.Panic = presentation.panicHandler

	return &presentation
}

func (r *Presentation) errorHandler(
	ctx *ext.Context,
	update *ext.Update,
	errorString string,
) error {
	zerolog.Ctx(ctx).Error().
		Err(errors.New(errorString)).
		Interface("u", update).
		Msg("error.while.processing.update")

	return nil
}

func (r *Presentation) panicHandler(
	ctx *ext.Context,
	update *ext.Update,
	errorString string,
) {
	zerolog.Ctx(ctx.Context).
		Error().
		Err(errors.New(errorString)).
		Interface("update", update).
		Msg("panic.while.processing.update")
}

func (r *Presentation) Run(ctx context.Context) error {
	user := r.protoClient.Self
	zerolog.Ctx(ctx).
		Info().
		Str("username", user.Username).
		Msg("starting.bot")

	err := r.protoClient.Idle()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
