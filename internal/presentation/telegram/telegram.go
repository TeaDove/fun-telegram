package telegram

import (
	"context"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/service/analitics"
	"github.com/teadove/goteleout/internal/supplier/ip_locator"
	"time"

	"github.com/teadove/goteleout/internal/supplier/kandinsky_supplier"
	"golang.org/x/time/rate"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/teadove/goteleout/internal/service/storage"
	"github.com/teadove/goteleout/internal/utils"
)

var (
	ErrBadUpdate    = errors.New("bad update")
	ErrPeerNotFound = errors.Wrap(ErrBadUpdate, "peer not found")
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
}

func MustNewTelegramPresentation(
	telegramAppID int,
	telegramAppHash string,
	telegramPhoneNumber string,
	sessionFullPath string,
	storage storage.Interface,
	kandinskySupplier *kandinsky_supplier.Supplier,
	ipLocator *ip_locator.Supplier,
	dbRepository *db_repository.Repository,
	analiticsService *analitics.Service,
) *Presentation {
	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		zerolog.Ctx(ctx).Warn().Str("status", "flood.waiting").Dur("dur", wait.Duration).Send()
	})

	middlewares := []telegram.Middleware{
		ratelimit.New(rate.Every(time.Millisecond*100), 10), waiter,
	}

	protoClient, err := gotgproto.NewClient(
		telegramAppID,
		telegramAppHash,
		gotgproto.ClientType{
			Phone: telegramPhoneNumber,
		},
		&gotgproto.ClientOpts{
			InMemory:         false,
			DisableCopyright: true,
			Session:          sessionMaker.SqliteSession(sessionFullPath),
			Middlewares:      middlewares,
			RunMiddleware: func(origRun func(ctx context.Context, f func(ctx context.Context) error) (err error), ctx context.Context, f func(ctx context.Context) (err error)) (err error) {
				return origRun(ctx, func(ctx context.Context) error {
					return waiter.Run(ctx, f)
				})
			},
		})
	utils.Check(err)

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
		"echo":          presentation.echoCommandHandler,
		"help":          presentation.helpCommandHandler,
		"get_me":        presentation.getMeCommandHandler,
		"ping":          presentation.pingCommandHandler,
		"spam_reaction": presentation.spamReactionCommandHandler,
		"kandinsky":     presentation.kandkinskyCommandHandler,
		"disable":       presentation.disableCommandHandler,
		"location":      presentation.locationCommandHandler,
		"stats":         presentation.statsCommandHandler,
		"upload_stats":  presentation.uploadStatsCommandHandler,
	}

	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback:      presentation.spamReactionMessageHandler,
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

func (r *Presentation) Run() error {
	err := r.protoClient.Idle()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
