package client

import (
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/supplier/telegram"
)

type Service struct {
	telegramSupplier               *telegram.Supplier
	telegramSessionStorageFullPath string
}

func MustNewClientService(telegramSupplier *telegram.Supplier, telegramSessionStorageFullPath string) Service {
	service := Service{telegramSupplier: telegramSupplier, telegramSessionStorageFullPath: telegramSessionStorageFullPath}

	return service
}

func (r Service) Run() {
	log.Info().
		Str("status", "starting.application").
		Msgf("All sessions will be stored in %s", r.telegramSessionStorageFullPath)
	err := r.telegramSupplier.Run()
	if err != nil {
		log.Panic().Stack().Err(err).Str("status", "error.while.running.telegram.client").Send()
	}
}
