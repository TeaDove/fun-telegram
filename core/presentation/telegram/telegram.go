package telegram

import (
	"context"
	"github.com/teadove/fun_telegram/core/repository/db_repository"
	"time"

	"github.com/teadove/fun_telegram/core/service/tex"

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
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"github.com/teadove/fun_telegram/core/repository/redis_repository"
	"github.com/teadove/fun_telegram/core/service/analitics"
	"github.com/teadove/fun_telegram/core/service/job"
	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/shared"
	"github.com/teadove/fun_telegram/core/supplier/ip_locator"
	"github.com/teadove/fun_telegram/core/supplier/kandinsky_supplier"
	"golang.org/x/time/rate"
)

type Presentation struct {
	// Unused, but may be usefully later
	//  telegramClient  *telegram.Client
	telegramApi     *tg.Client
	telegramManager *peers.Manager
	protoClient     *gotgproto.Client

	router   map[string]messageProcessor
	features map[string]bool

	kandinskySupplier *kandinsky_supplier.Supplier
	ipLocator         *ip_locator.Supplier
	redisRepository   *redis_repository.Repository
	mongoRepository   *mongo_repository.Repository
	dbRepository      *db_repository.Repository
	resourceService   *resource.Service
	analiticsService  *analitics.Service
	jobService        *job.Service
	texService        *tex.Service
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
		// WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		//	zerolog.Ctx(ctx).
		//		Warn().
		//		Str("status", "flood.waiting").
		//		Str("dur", wait.Duration.String()).
		//		Send()
		// })
		//runMiddleware = func(origRun func(ctx context.Context, f func(ctx context.Context) error) (err error), ctx context.Context, f func(ctx context.Context) (err error)) (err error) {
		//	return origRun(ctx, func(ctx context.Context) error {
		//		return waiter.Run(ctx, f)
		//	})
		//}

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
	kandinskySupplier *kandinsky_supplier.Supplier,
	ipLocator *ip_locator.Supplier,
	mongoRepository *mongo_repository.Repository,
	analiticsService *analitics.Service,
	jobService *job.Service,
	resourceService *resource.Service,
	dbRepository *db_repository.Repository,
) *Presentation {
	api := protoClient.API()

	presentation := Presentation{
		redisRepository:   redisRepository,
		protoClient:       protoClient,
		telegramApi:       api,
		telegramManager:   peers.Options{}.Build(api),
		kandinskySupplier: kandinskySupplier,
		ipLocator:         ipLocator,
		mongoRepository:   mongoRepository,
		analiticsService:  analiticsService,
		jobService:        jobService,
		resourceService:   resourceService,
		dbRepository:      dbRepository,
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
		"echo": {
			executor:    presentation.echoCommandHandler,
			description: resource.CommandEchoDescription,
			flags:       []optFlag{},
			example:     "Hello World!",
		},
		"help": {
			executor:    presentation.helpCommandHandler,
			description: resource.CommandHelpDescription,
			flags:       []optFlag{},
		},
		"get_me": {
			executor:    presentation.getMeCommandHandler,
			description: resource.CommandGetMeHelpDescription,
			flags:       []optFlag{},
		},
		"ping": {
			executor:          presentation.pingCommandHandler,
			description:       resource.CommandPingDescription,
			flags:             []optFlag{},
			requireAdmin:      true,
			disabledByDefault: true,
		},
		"spam_reaction": {
			executor:          presentation.spamReactionCommandHandler,
			description:       resource.CommandSpamReactionDescription,
			flags:             []optFlag{FlagSpamReactionStop},
			disabledByDefault: true,
		},
		"kandinsky": {
			executor:    presentation.kandkinskyCommandHandler,
			description: resource.CommandKandinskyDescription,
			flags: []optFlag{
				FlagKandinskyNegativePrompt,
				FlagKandinskyStyle,
				FlagKandinskyPageStyle,
				FlagKandinskyCountStyle,
			},
			example: "-c=3 --style=ANIME girl in space, sticker, realism, cute_mood, bold colors, disney",
		},
		"regrule": {
			executor:     presentation.regruleCommandHandler,
			description:  resource.CommandRegRuleDescription,
			flags:        []optFlag{FlagRegRuleRegexp, FlagRegRuleDelete, FlagRegRuleList},
			example:      "—regexp=\"^\\w+$\" ПРИВЕТ ЧЕ КАК",
			requireAdmin: true,
		},
		"location": {
			executor:    presentation.locationCommandHandler,
			description: resource.CommandLocationDescription,
			flags:       []optFlag{},
		},
		"stats": {
			executor:    presentation.statsCommandHandler,
			description: resource.CommandStatsDescription,
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
			description: resource.CommandUploadStatsDescription,
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
			description:  resource.CommandUploadStatsDescription,
			requireOwner: true,
			flags: []optFlag{
				FlagStatsChannelName,
				FlagStatsChannelDepth,
				FlagStatsChannelMaxOrder,
			},
		},
		"ban": {
			executor:    presentation.banCommandHandler,
			description: resource.CommandBanDescription,
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
		"chat": {
			executor:     presentation.chatCommandHandler,
			description:  resource.CommandChatDescription,
			requireAdmin: true,
			flags:        []optFlag{FlagChatTz, FlagChatLocale, FlagChatEnabled},
			example:      "-e -l=ru -t=3",
		},
		"restart": {
			executor:     presentation.restartCommandHandler,
			description:  resource.CommandRestartDescription,
			requireOwner: true,
		},
		"anime-detect": {
			executor:    presentation.animeDetectionCommandHandler,
			description: resource.CommandAnimeDetectDescription,
		},
		"tex": {
			executor:    presentation.texCommandHandler,
			description: resource.CommandTexDescription,
			example:     "Найс! $f(x) = \\frac{\\sqrt{x +20}}{2\\pi} +\\hbar \\sum y\\partial y$",
		},
	}

	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.spamReactionMessageHandler,
			Filters:       filters.Message.Text,
			UpdateFilters: filterNonNewMessagesNotFromUser,
			Outgoing:      true,
		},
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.regRuleFinderMessagesProcessor,
			Filters:       filters.Message.Text,
			UpdateFilters: filterNonNewMessagesNotFromUser,
			Outgoing:      true,
		},
	)
	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.animeDetectionMessagesProcessor,
			Outgoing:      true,
			Filters:       filters.Message.Media,
			UpdateFilters: filterNonNewMessagesNotFromUser,
		},
	)

	dp, ok := protoClient.Dispatcher.(*dispatcher.NativeDispatcher)
	if !ok {
		shared.FancyPanic(ctx, errors.New("can only work with NativeDispatcher"))
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

	_, err = scheduler.
		Every(1*time.Minute).
		Do(shared.CheckOfLog(presentation.deleteOldPingMessages), ctx)
	shared.Check(ctx, err)

	scheduler.StartAsync()

	presentation.setFeatures()

	zerolog.Ctx(ctx).Info().
		Str("status", "telegram.presentation.created").
		Dur("retry.interval", presentation.protoClient.RetryInterval).
		Int("retry.count", presentation.protoClient.MaxRetries).
		Send()

	return &presentation
}

func (r *Presentation) setFeatures() {
	r.features = make(map[string]bool, len(r.router))
	for commandName, command := range r.router {
		r.features[commandName] = !command.disabledByDefault
	}

	r.features[animeDetectionFeatureName] = false
}

func (r *Presentation) errorHandler(
	ctx *ext.Context,
	update *ext.Update,
	errorString string,
) error {
	zerolog.Ctx(ctx).Error().
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
		Str("status", "panic.while.processing.update").
		Interface("update", update).
		Send()
	println(errorString)
}

func (r *Presentation) Run() error {
	err := r.protoClient.Idle()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
