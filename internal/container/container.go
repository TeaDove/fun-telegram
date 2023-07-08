package container

import (
	"github.com/teadove/goteleout/internal/service/storage/memory"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/presentation/telegram"
	"github.com/teadove/goteleout/internal/service/client"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/utils"
)

type Container struct {
	Presentation                   *telegram.Presentation
	TelegramSessionStorageFullPath string
}

func MustNewCombatContainer() Container {
	settings := shared.MustNewSettings()
	level, err := zerolog.ParseLevel(settings.LogLevel)
	utils.Check(err)

	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	realPath, err := homedir.Expand(settings.FileStoragePath)
	utils.Check(err)
	err = os.MkdirAll(realPath, os.ModePerm)
	utils.Check(err)

	clientService := client.MustNewClientService()
	memoryStorage := memory.MustNew()

	telegramPresentation := telegram.MustNewTelegramPresentation(
		&clientService,
		settings.Telegram.AppID,
		settings.Telegram.AppHash,
		settings.Telegram.PhoneNumber,
		settings.Telegram.SessionFullPath,
		memoryStorage,
	)

	container := Container{&telegramPresentation, settings.Telegram.SessionFullPath}
	return container
}
