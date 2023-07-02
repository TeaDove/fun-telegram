package telegram

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/service/client"
)

type Presentation struct {
	telegramClient *telegram.Client

	clientService *client.Service
}

func MustNewTelegramPresentation(clientService *client.Service, telegramAppID int, telegramAppHash string, telegramSessionStorageFullPath string) Presentation {
	// https://core.telegram.org/api/obtaining_api_id

	sessionStorage := telegram.FileSessionStorage{Path: telegramSessionStorageFullPath}
	telegramClient := telegram.NewClient(telegramAppID, telegramAppHash, telegram.Options{SessionStorage: &sessionStorage})

	service := Presentation{telegramClient: telegramClient, clientService: clientService}
	return service
}

func (r Presentation) login(ctx context.Context) error {
	flow := auth.NewFlow(terminalAuth{}, auth.SendCodeOptions{})
	status, err := r.telegramClient.Auth().Status(ctx)
	if !status.Authorized {
		log.Info().Str("status", "authorizing").Send()
		err = r.telegramClient.Auth().IfNecessary(ctx, flow)
	}
	if err != nil {
		return errors.Join(err, errors.New("error while authenticating"))
	}

	return nil
}

func (r Presentation) Run() error {

	return r.telegramClient.Run(context.Background(), func(ctx context.Context) error {
		api := r.telegramClient.API()
		err := r.login(ctx)
		if err != nil {
			return err
		}

		s := message.NewSender(api)
		_, err = s.Self().Text(ctx, "Hi!")
		if err != nil {
			return err
		}

		return nil
	})
}
