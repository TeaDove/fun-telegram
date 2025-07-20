package telegram

import (
	"context"
	"time"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/glebarez/sqlite"

	"github.com/go-co-op/gocron"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/redis_repository"
	"github.com/teadove/fun_telegram/core/service/analitics"
	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/shared"
	"golang.org/x/time/rate"
)

type Presentation struct {
	telegramApi     *tg.Client
	telegramManager *peers.Manager
	protoClient     *gotgproto.Client

	router   map[string]messageProcessor
	features map[string]bool

	redisRepository  *redis_repository.Repository
	dbRepository     *db_repository.Repository
	resourceService  *resource.Service
	analiticsService *analitics.Service
}

func NewProtoClient(ctx context.Context) (*gotgproto.Client, error) {
	middlewares := make([]telegram.Middleware, 0, 2)

	if shared.AppSettings.Telegram.RateLimiterEnabled {
		middlewares = append(middlewares, ratelimit.New(
			rate.Every(shared.AppSettings.Telegram.RateLimiterRate),
			shared.AppSettings.Telegram.RateLimiterLimit))
	}

	var runMiddleware func(
		origRun func(ctx context.Context, f func(ctx context.Context) error) (err error),
		ctx context.Context,
		f func(ctx context.Context) (err error),
	) (err error)

	if shared.AppSettings.Telegram.FloodWaiterEnabled {
		waiter := floodwait.
			NewSimpleWaiter().
			WithMaxWait(time.Minute * 10).
			WithMaxRetries(20)

		middlewares = append(middlewares, waiter)

	}

	protoClient, err := gotgproto.NewClient(
		shared.AppSettings.Telegram.AppID,
		shared.AppSettings.Telegram.AppHash,
		gotgproto.ClientTypePhone(shared.AppSettings.Telegram.PhoneNumber),
		&gotgproto.ClientOpts{
			Context:          ctx,
			InMemory:         false,
			DisableCopyright: true,
			Session: sessionMaker.SqlSession(
				sqlite.Open(".mtproto"),
			),
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
	ctx context.Context,
	protoClient *gotgproto.Client,
	redisRepository *redis_repository.Repository,
	analiticsService *analitics.Service,
	resourceService *resource.Service,
	dbRepository *db_repository.Repository,
) *Presentation {
	api := protoClient.API()

	presentation := Presentation{
		redisRepository:  redisRepository,
		protoClient:      protoClient,
		telegramApi:      api,
		telegramManager:  peers.Options{}.Build(api),
		analiticsService: analiticsService,
		resourceService:  resourceService,
		dbRepository:     dbRepository,
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
		"get_me": {
			executor:    presentation.getMeCommandHandler,
			description: "get id, username of requested user and group",
			flags:       []optFlag{},
		},
		"stats": {
			executor:    presentation.statsCommandHandler,
			description: "get stats of this chat",
			flags: []optFlag{
				FlagStatsUsername,
				FlagStatsChannelName,
				FlagStatsChannelDepth,
				FlagStatsChannelMaxOrder,
				FlagStatsAnonymize,
			},
			requireAdmin: true,
		},
		"upload_stats": {
			executor:    presentation.uploadStatsCommandHandler,
			description: "uploads stats from this chat",
			flags: []optFlag{
				FlagUploadStatsRemove,
				FlagUploadStatsCount,
				FlagUploadStatsDay,
				FlagUploadStatsOffset,
				FlagStatsChannelName,
				FlagStatsChannelDepth,
				FlagStatsChannelMaxOrder,
			},
			requireAdmin: true,
			example:      "-c=400000 -d=365 -o=0 --silent",
		},
		"dump_stats": {
			executor:     presentation.statsDumpCommandHandler,
			description:  "uploads stats from this chat",
			requireOwner: true,
			flags: []optFlag{
				FlagStatsChannelName,
				FlagStatsChannelDepth,
				FlagStatsChannelMaxOrder,
			},
		},
		"ban": {
			executor:    presentation.banCommandHandler,
			description: "bans or unbans user from using this bot globally",
		},

		"restart": {
			executor:     presentation.restartCommandHandler,
			description:  "restarts bot",
			requireOwner: true,
		},
	}

	dp, ok := protoClient.Dispatcher.(*dispatcher.NativeDispatcher)
	if !ok {
		panic("telegram dispatcher is not native")
	}

	dp.Error = presentation.errorHandler
	dp.Panic = presentation.panicHandler

	err := presentation.updateRestartMessages(ctx)
	if err != nil {
		zerolog.Ctx(ctx).
			Error().
			Stack().
			Err(err).
			Str("status", "failed.to.update.restart.messages").
			Send()
	}

	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.StartAsync()

	zerolog.Ctx(ctx).Info().
		Msg("telegram.presentation.created")

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

func (r *Presentation) Run() error {
	err := r.protoClient.Idle()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
