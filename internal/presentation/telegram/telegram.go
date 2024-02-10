package telegram

import (
	"context"
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
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"github.com/teadove/goteleout/internal/repository/redis_repository"
	"github.com/teadove/goteleout/internal/service/analitics"
	"github.com/teadove/goteleout/internal/service/job"
	"github.com/teadove/goteleout/internal/service/resource"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/supplier/ip_locator"
	"github.com/teadove/goteleout/internal/supplier/kandinsky_supplier"
	"github.com/teadove/goteleout/internal/utils"
	"golang.org/x/time/rate"
	"time"
)

type Presentation struct {
	// Unused, but may be usefully later
	//  telegramClient  *telegram.Client
	telegramApi     *tg.Client
	telegramManager *peers.Manager
	protoClient     *gotgproto.Client

	router map[string]messageProcessor

	kandinskySupplier *kandinsky_supplier.Supplier
	ipLocator         *ip_locator.Supplier
	redisRepository   *redis_repository.Repository
	dbRepository      *mongo_repository.Repository
	resourceService   *resource.Service
	analiticsService  *analitics.Service
	jobService        *job.Service
}

func MustNewProtoClient(ctx context.Context) *gotgproto.Client {
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
			NewWaiter().
			WithMaxWait(time.Minute * 5).
			WithMaxRetries(20).
			WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
				zerolog.Ctx(ctx).Warn().Str("status", "flood.waiting").Dur("dur", wait.Duration).Send()
			})

		middlewares = append(middlewares, waiter)
		runMiddleware = func(origRun func(ctx context.Context, f func(ctx context.Context) error) (err error), ctx context.Context, f func(ctx context.Context) (err error)) (err error) {
			return origRun(ctx, func(ctx context.Context) error {
				return waiter.Run(ctx, f)
			})
		}

	}

	protoClient, err := gotgproto.NewClient(
		shared.AppSettings.Telegram.AppID,
		shared.AppSettings.Telegram.AppHash,
		gotgproto.ClientType{
			Phone: shared.AppSettings.Telegram.PhoneNumber,
		},
		&gotgproto.ClientOpts{
			Context:          ctx,
			InMemory:         false,
			DisableCopyright: true,
			Session:          sessionMaker.SqliteSession(shared.AppSettings.Telegram.SessionFullPath),
			Middlewares:      middlewares,
			RunMiddleware:    runMiddleware,
		})
	utils.Check(err)

	return protoClient
}

func MustNewTelegramPresentation(
	ctx context.Context,
	protoClient *gotgproto.Client,
	redisRepository *redis_repository.Repository,
	kandinskySupplier *kandinsky_supplier.Supplier,
	ipLocator *ip_locator.Supplier,
	dbRepository *mongo_repository.Repository,
	analiticsService *analitics.Service,
	jobService *job.Service,
	resourceService *resource.Service,
) *Presentation {
	api := protoClient.API()

	presentation := Presentation{
		redisRepository:   redisRepository,
		protoClient:       protoClient,
		telegramApi:       api,
		telegramManager:   peers.Options{}.Build(api),
		kandinskySupplier: kandinskySupplier,
		ipLocator:         ipLocator,
		dbRepository:      dbRepository,
		analiticsService:  analiticsService,
		jobService:        jobService,
		resourceService:   resourceService,
	}

	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback: presentation.injectContext, Outgoing: true,
		},
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback: presentation.catchMessages, Outgoing: true,
		},
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback: presentation.deleteOut, Outgoing: true,
		},
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback: presentation.route, Outgoing: true,
		},
	)

	presentation.router = map[string]messageProcessor{
		"echo": {
			executor:    presentation.echoCommandHandler,
			description: resource.CommandEchoHelp,
			flags:       []OptFlag{},
		},
		"help": {
			executor:    presentation.helpCommandHandler,
			description: resource.CommandHelpDescription,
			flags:       []OptFlag{},
		},
		"get_me": {
			executor:    presentation.getMeCommandHandler,
			description: resource.CommandGetMeHelpDescription,
			flags:       []OptFlag{},
		},
		"ping": {
			executor:     presentation.pingCommandHandler,
			description:  resource.CommandPingDescription,
			flags:        []OptFlag{},
			requireAdmin: true,
		},
		"spam_reaction": {
			executor:    presentation.spamReactionCommandHandler,
			description: resource.CommandSpamReactionDescription,
			flags:       []OptFlag{FlagStop},
		},
		"kandinsky": {
			executor:    presentation.kandkinskyCommandHandler,
			description: resource.CommandKandinskyDescription,
			flags:       []OptFlag{FlagNegativePrompt, FlagStyle},
		},
		"disable": {
			executor:     presentation.disableCommandHandler,
			description:  resource.CommandDisableDescription,
			flags:        []OptFlag{},
			requireAdmin: true,
		},
		"location": {
			executor:    presentation.locationCommandHandler,
			description: resource.CommandLocationDescription,
			flags:       []OptFlag{},
		},
		"stats": {
			executor:     presentation.statsCommandHandler,
			description:  resource.CommandStatsDescription,
			flags:        []OptFlag{FlagTZ, FlagStatsUsername},
			requireAdmin: true,
		},
		"upload_stats": {
			executor:     presentation.uploadStatsCommandHandler,
			description:  resource.CommandUploadStatsDescription,
			flags:        []OptFlag{FlagRemove, FlagCount, FlagDay},
			requireAdmin: true,
		},
		"ban": {
			executor:    presentation.banCommandHandler,
			description: resource.CommandBanDescription,
		},
		"toxic": {
			executor:     presentation.toxicFinderCommandHandler,
			description:  resource.CommandToxicDescription,
			requireAdmin: true,
		},
		"health": {
			executor:    presentation.healthCommandHandler,
			description: resource.CommandHealthDescription,
		},
		"infra_stats": {
			executor:     presentation.infraStatsCommandHandler,
			description:  resource.CommandInfraStatsDescription,
			requireOwner: true,
		},
		"locale": {
			executor:     presentation.localeCommandHandler,
			description:  resource.CommandLocaleDescription,
			requireAdmin: true,
		},
		"reload": {
			executor:     presentation.restartCommandHandler,
			description:  resource.CommandLocaleDescription,
			requireOwner: true,
		},
	}

	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.spamReactionMessageHandler,
			Filters:       nil,
			UpdateFilters: nil,
			Outgoing:      true,
		},
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.toxicFinderMessagesProcessor,
			Filters:       nil,
			UpdateFilters: nil,
			Outgoing:      true,
		},
	)

	dp, ok := protoClient.Dispatcher.(*dispatcher.NativeDispatcher)
	if !ok {
		utils.FancyPanic(errors.New("can only work with NativeDispatcher"))
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
	zerolog.Ctx(ctx.Context).Error().
		Stack().
		Err(errors.New(errorString)).
		Str("status", "error.while.processing.update").
		Interface("update", update).
		Send()

	return nil
}

func (r *Presentation) panicHandler(
	ctx *ext.Context,
	update *ext.Update,
	errorString string,
) {
	zerolog.Ctx(ctx.Context).Error().
		Stack().
		Err(errors.New(errorString)).
		Str("status", "panic.while.processing.update").
		Interface("update", update).
		Send()
}

func (r *Presentation) Run() error {
	err := r.protoClient.Idle()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
