package telegram

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/service/client"
	"strings"
)

type Presentation struct {
	telegramClient     *telegram.Client
	telegramDispatcher *tg.UpdateDispatcher
	telegramSender     *message.Sender
	telegramApi        *tg.Client
	commandHandler     map[string]commandProcessor

	clientService *client.Service
}

func MustNewTelegramPresentation(clientService *client.Service, telegramAppID int, telegramAppHash string, telegramSessionStorageFullPath string) Presentation {
	// https://core.telegram.org/api/obtaining_api_id

	sessionStorage := telegram.FileSessionStorage{Path: telegramSessionStorageFullPath}
	updater := tg.NewUpdateDispatcher()
	telegramClient := telegram.NewClient(telegramAppID, telegramAppHash, telegram.Options{SessionStorage: &sessionStorage, UpdateHandler: updater})
	api := telegramClient.API()
	sender := message.NewSender(api)

	presentation := Presentation{telegramClient: telegramClient, clientService: clientService, telegramDispatcher: &updater, telegramApi: api, telegramSender: sender}
	presentation.commandHandler = map[string]commandProcessor{
		"ping": presentation.pingCommandHandler,
		"help": presentation.helpCommandHandler}

	return presentation
}

var (
	BadUpdate = errors.New("bad update")
)

func (r *Presentation) login(ctx context.Context) error {
	flow := auth.NewFlow(terminalAuth{}, auth.SendCodeOptions{})
	status, err := r.telegramClient.Auth().Status(ctx)
	if !status.Authorized {
		log.Info().Str("status", "authorizing").Send()
		err = r.telegramClient.Auth().IfNecessary(ctx, flow)
	}
	if err != nil {
		return errors.Join(err, errors.New("error while authenticating"))
	}

	_, err = r.telegramSender.Self().Text(ctx, "Telegram client initialized")
	if err != nil {
		return err
	}

	return nil
}

func (r *Presentation) routeMessage(ctx context.Context, entities *tg.Entities, update message.AnswerableMessageUpdate) error {
	m, ok := update.GetMessage().(*tg.Message)
	if !ok {
		return BadUpdate
	}

	if m.Post {
		return nil
	}
	log.Debug().Str("status", "message.got").Str("text", m.Message).Interface("message", m).Send()

	firstMessage := strings.Split(m.Message, " ")[0]
	const commandPrefix = '!'
	if len(firstMessage) < 1 {
		return nil
	}
	if firstMessage[0] != commandPrefix {
		return nil
	}
	command := firstMessage[1:]
	log.Info().Str("status", "command.got").Str("command", command).Send()
	handler, ok := r.commandHandler[command]
	if !ok {
		log.Info().Str("status", "unknown.command").Str("command", command).Str("text", m.Message).Send()
		return nil
	}
	return handler(ctx, entities, update, m)
}

func (r *Presentation) Run() error {
	return r.telegramClient.Run(context.Background(), func(ctx context.Context) error {
		err := r.login(ctx)
		if err != nil {
			return err
		}

		r.telegramDispatcher.OnNewChannelMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewChannelMessage) error {
			return r.routeMessage(ctx, &entities, update)
		})

		r.telegramDispatcher.OnNewMessage(func(ctx context.Context, entities tg.Entities, update *tg.UpdateNewMessage) error {
			return r.routeMessage(ctx, &entities, update)
		})
		err = telegram.RunUntilCanceled(context.Background(), r.telegramClient)
		if err != nil {
			return err
		}

		return nil
	})
}
