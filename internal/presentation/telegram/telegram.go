package telegram

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/supplier/kandinsky_supplier"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/teadove/goteleout/internal/service/client"
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
	telegramSender  *message.Sender
	telegramApi     *tg.Client
	telegramManager *peers.Manager
	protoClient     *gotgproto.Client

	storage           storage.Interface
	clientService     *client.Service
	router            map[string]messageProcessor
	kandinskySupplier *kandinsky_supplier.Supplier

	logErrorToSelf bool
}

func MustNewTelegramPresentation(
	clientService *client.Service,
	telegramAppID int,
	telegramAppHash string,
	telegramPhoneNumber string,
	sessionFullPath string,
	storage storage.Interface,
	logErrorToSelf bool,
	kandinskySupplier *kandinsky_supplier.Supplier,
) Presentation {
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
		})
	utils.Check(err)

	api := protoClient.API()

	presentation := Presentation{
		clientService:     clientService,
		storage:           storage,
		protoClient:       protoClient,
		telegramApi:       api,
		telegramSender:    message.NewSender(api),
		telegramManager:   peers.Options{}.Build(api),
		logErrorToSelf:    logErrorToSelf,
		kandinskySupplier: kandinskySupplier,
	}

	protoClient.Dispatcher.AddHandler(
		handlers.Message{
			Callback: presentation.injectContext, Outgoing: true,
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

	return presentation
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

	if r.logErrorToSelf {
		_, err := ctx.SendMessage(ctx.Self.ID, &tg.MessagesSendMessageRequest{
			Silent:  true,
			Message: fmt.Sprintf("Error occurred while processing update:\n\n%s", errorString),
		})
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	return nil
}

func (r *Presentation) Run() error {
	ctx := r.protoClient.CreateContext()

	_, err := ctx.SendMessage(
		ctx.Self.ID,
		&tg.MessagesSendMessageRequest{Message: "Fun telegram initialized!"},
	)
	if err != nil {
		return err
	}

	err = r.protoClient.Idle()

	return err
}
