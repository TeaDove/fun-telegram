package telegramapi

import (
	"context"
	"time"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/glebarez/sqlite"

	"github.com/TeaDove/fun-telegram/internal/services/analyticsservice"
	"github.com/TeaDove/fun-telegram/internal/utils"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/cockroachdb/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

type Presentation struct {
	telegramAPI     *tg.Client
	telegramManager *peers.Manager
	protoClient     *gotgproto.Client

	router map[string]messageProcessor

	analiticsService *analyticsservice.Service
}

func NewProtoClient(ctx context.Context) (*gotgproto.Client, error) {
	middlewares := make([]telegram.Middleware, 0, 2)

	if utils.AppSettings.Telegram.RateLimiterEnabled {
		middlewares = append(middlewares, ratelimit.New(
			rate.Every(utils.AppSettings.Telegram.RateLimiterRate),
			utils.AppSettings.Telegram.RateLimiterLimit))
	}

	var runMiddleware func(
		origRun func(ctx context.Context, f func(ctx context.Context) error) (err error),
		ctx context.Context,
		f func(ctx context.Context) (err error),
	) (err error)

	if utils.AppSettings.Telegram.FloodWaiterEnabled {
		waiter := floodwait.
			NewSimpleWaiter().
			WithMaxWait(time.Minute * 10).
			WithMaxRetries(20)

		middlewares = append(middlewares, waiter)
	}

	protoClient, err := gotgproto.NewClient(
		utils.AppSettings.Telegram.AppID,
		utils.AppSettings.Telegram.AppHash,
		gotgproto.ClientTypePhone(utils.AppSettings.Telegram.PhoneNumber),
		&gotgproto.ClientOpts{
			Context:          ctx,
			InMemory:         false,
			DisableCopyright: true,
			Session:          sessionMaker.SqlSession(sqlite.Open(".mtproto")),
			AuthConversator: authConversator{
				AuthConversator: gotgproto.BasicConversator(),
				logger:          zerolog.Ctx(ctx),
			},
			Middlewares:   middlewares,
			RunMiddleware: runMiddleware,
			RetryInterval: 10 * time.Second,
			MaxRetries:    10,
			DC:            2,
		})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gotgproto client")
	}

	return protoClient, nil
}

func MustNewTelegramPresentation(
	protoClient *gotgproto.Client,
	analiticsService *analyticsservice.Service,
) *Presentation {
	api := protoClient.API()

	presentation := Presentation{
		protoClient:      protoClient,
		telegramAPI:      api,
		telegramManager:  peers.Options{}.Build(api),
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
			Callback:      presentation.deleteOut,
			Filters:       filters.Message.Text,
			UpdateFilters: filterNonNewMessagesNotFromUser,
			Outgoing:      true,
		},
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.route,
			Filters:       filters.Message.Text,
			UpdateFilters: filterNonNewMessagesNotFromUser,
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
			description: "uploads stats from this chat",
			flags: []optFlag{
				FlagUploadStatsCount,
				FlagUploadStatsDay,
				FlagUploadStatsOffset,
				FlagStatsAnonymize,
			},
			example: "-c=400000 -d=365 -o=0 --silent",
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
