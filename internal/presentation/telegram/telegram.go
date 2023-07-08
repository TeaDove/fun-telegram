package telegram

import (
	"errors"
	"github.com/anonyindian/gotgproto"
	"github.com/anonyindian/gotgproto/dispatcher/handlers"
	"github.com/anonyindian/gotgproto/sessionMaker"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/teadove/goteleout/internal/service/client"
	"github.com/teadove/goteleout/internal/service/storage"
	"github.com/teadove/goteleout/internal/utils"
)

var BadUpdate = errors.New("bad update")

type Presentation struct {
	telegramClient  *telegram.Client
	telegramSender  *message.Sender
	telegramApi     *tg.Client
	telegramManager *peers.Manager
	protoClient     *gotgproto.Client

	storage       storage.Interface
	clientService *client.Service
}

func MustNewTelegramPresentation(
	clientService *client.Service,
	telegramAppID int,
	telegramAppHash string,
	telegramPhoneNumber string,
	telegramSessionStorageFullPath string,
	storage storage.Interface,
) Presentation {

	protoClient, err := gotgproto.NewClient(telegramAppID, telegramAppHash, gotgproto.ClientType{
		Phone: telegramPhoneNumber,
	}, &gotgproto.ClientOpts{
		DisableCopyright: true,
		Session:          sessionMaker.NewSession(telegramSessionStorageFullPath, sessionMaker.Session),
	})
	utils.Check(err)

	api := protoClient.API()

	presentation := Presentation{
		clientService:   clientService,
		storage:         storage,
		protoClient:     protoClient,
		telegramApi:     api,
		telegramSender:  message.NewSender(api),
		telegramManager: peers.Options{}.Build(api),
	}

	protoClient.Dispatcher.AddHandler(handlers.NewCommand("echo", presentation.echoCommandHandler))
	protoClient.Dispatcher.AddHandler(handlers.NewCommand("help", presentation.helpCommandHandler))
	protoClient.Dispatcher.AddHandler(handlers.NewCommand("getMe", presentation.getMeCommandHandler))
	protoClient.Dispatcher.AddHandler(handlers.NewCommand("ping", presentation.pingCommandHandler))
	protoClient.Dispatcher.AddHandler(handlers.NewCommand("spamReaction", presentation.spamReactionCommandHandler))

	return presentation
}

func (r *Presentation) Run() error {
	ctx := r.protoClient.CreateContext()
	_, err := ctx.SendMessage(ctx.Self.ID, &tg.MessagesSendMessageRequest{Message: "Fun telegram initialized!"})
	if err != nil {
		return err
	}

	err = r.protoClient.Idle()
	return err
}
