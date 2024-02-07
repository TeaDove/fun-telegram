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
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/service/analitics"
	"github.com/teadove/goteleout/internal/service/job"
	"github.com/teadove/goteleout/internal/service/storage"
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

	storage storage.Interface
	router  map[string]messageProcessor

	kandinskySupplier *kandinsky_supplier.Supplier
	ipLocator         *ip_locator.Supplier

	dbRepository     *db_repository.Repository
	analiticsService *analitics.Service
	helpMessage      []styling.StyledTextOption
	jobService       *job.Service
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

	//zapLogger, err := zap.NewDevelopment()
	//zapLogger.WithOptions()
	//utils.Check(err)

	protoClient, err := gotgproto.NewClient(
		shared.AppSettings.Telegram.AppID,
		shared.AppSettings.Telegram.AppHash,
		gotgproto.ClientType{
			Phone: shared.AppSettings.Telegram.PhoneNumber,
		},
		&gotgproto.ClientOpts{
			//Logger:           zapLogger,
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
	storage storage.Interface,
	kandinskySupplier *kandinsky_supplier.Supplier,
	ipLocator *ip_locator.Supplier,
	dbRepository *db_repository.Repository,
	analiticsService *analitics.Service,
	jobService *job.Service,
) *Presentation {
	api := protoClient.API()

	presentation := Presentation{
		storage:           storage,
		protoClient:       protoClient,
		telegramApi:       api,
		telegramManager:   peers.Options{}.Build(api),
		kandinskySupplier: kandinskySupplier,
		ipLocator:         ipLocator,
		dbRepository:      dbRepository,
		analiticsService:  analiticsService,
		jobService:        jobService,
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
			description: "echoes with same message",
			flags:       []tgUtils.OptFlag{},
		},
		"help": {
			executor:    presentation.helpCommandHandler,
			description: "get this message",
			flags:       []tgUtils.OptFlag{},
		},
		"get_me": {
			executor:    presentation.getMeCommandHandler,
			description: "get id, username of requested user and group",
			flags:       []tgUtils.OptFlag{},
		},
		"ping": {
			executor:     presentation.pingCommandHandler,
			description:  "ping all users",
			flags:        []tgUtils.OptFlag{},
			requireAdmin: true,
		},
		"spam_reaction": {
			executor:    presentation.spamReactionCommandHandler,
			description: "if replied to message with reaction, will spam this reaction to replied user",
			flags:       []tgUtils.OptFlag{FlagStop},
		},
		"kandinsky": {
			executor:    presentation.kandkinskyCommandHandler,
			description: "generate image via kandinsky",
			flags:       []tgUtils.OptFlag{FlagNegativePrompt, FlagStyle},
		},
		"disable": {
			executor:     presentation.disableCommandHandler,
			description:  "disables or enabled bot in this chat",
			flags:        []tgUtils.OptFlag{},
			requireAdmin: true,
		},
		"location": {
			executor:    presentation.locationCommandHandler,
			description: "get description by ip address or domain",
			flags:       []tgUtils.OptFlag{},
		},
		"stats": {
			executor:     presentation.statsCommandHandler,
			description:  "returns stats of this chat",
			flags:        []tgUtils.OptFlag{FlagTZ, FlagStatsUsername},
			requireAdmin: true,
		},
		"upload_stats": {
			executor:     presentation.uploadStatsCommandHandler,
			description:  "uploads stats from this chat",
			flags:        []tgUtils.OptFlag{FlagRemove, FlagCount, FlagDay},
			requireAdmin: true,
		},
		"ban": {
			executor:    presentation.banCommandHandler,
			description: "bans or unbans user from using this bot globally",
		},
		"toxic": {
			executor:     presentation.toxicFinderCommandHandler,
			description:  "find toxic words and screem about them",
			requireAdmin: true,
		},
		"health": {
			executor:    presentation.healthCommandHandler,
			description: "checks if server is not down",
		},
		"infra_stats": {
			executor:     presentation.infraStatsCommandHandler,
			description:  "show infrastraction load information",
			requireOwner: true,
		},
	}
	presentation.setHelpMessage()

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

func filterNonNewMessages(update *ext.Update) bool {
	switch update.UpdateClass.(type) {
	case *tg.UpdateNewChannelMessage:
		return true
	case *tg.UpdateNewMessage:
		return true
	default:
		return false
	}
}
