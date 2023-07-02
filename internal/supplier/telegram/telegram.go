package telegram

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/rs/zerolog/log"
)

type Supplier struct {
	client *telegram.Client
}

func MustNewTelegramSupplier(telegramAppID int, telegramAppHash string, telegramSessionStorageFullPath string) Supplier {
	// https://core.telegram.org/api/obtaining_api_id

	sessionStorage := telegram.FileSessionStorage{Path: telegramSessionStorageFullPath}
	client := telegram.NewClient(telegramAppID, telegramAppHash, telegram.Options{SessionStorage: &sessionStorage})

	service := Supplier{client: client}
	return service
}

func (r Supplier) login(ctx context.Context) error {
	flow := auth.NewFlow(terminalAuth{}, auth.SendCodeOptions{})
	status, err := r.client.Auth().Status(ctx)
	if !status.Authorized {
		log.Info().Str("status", "authorizing").Send()
		err = r.client.Auth().IfNecessary(ctx, flow)
	}
	if err != nil {
		return errors.Join(err, errors.New("error while authenticating"))
	}

	return nil
}

func (r Supplier) Run() error {
	return r.client.Run(context.Background(), func(ctx context.Context) error {
		api := r.client.API()
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
