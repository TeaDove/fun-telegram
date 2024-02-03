package container

import (
	"context"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/service/analitics"
	"github.com/teadove/goteleout/internal/service/job"
	"github.com/teadove/goteleout/internal/service/storage"
	"github.com/teadove/goteleout/internal/service/storage/redis"
	"github.com/teadove/goteleout/internal/supplier/ip_locator"
	"github.com/teadove/goteleout/internal/supplier/kandinsky_supplier"
	"os"
	"path/filepath"

	"github.com/teadove/goteleout/internal/service/storage/memory"

	"github.com/mitchellh/go-homedir"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/presentation/telegram"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/utils"
)

type Container struct {
	Presentation *telegram.Presentation
	JobService   *job.Service
}

func makeStorage(settings *shared.Settings) storage.Interface {
	realPath, err := homedir.Expand(settings.FileStoragePath)
	utils.Check(err)
	err = os.MkdirAll(realPath, os.ModePerm)
	utils.Check(err)

	path := filepath.Join(realPath, settings.Storage.Filename)

	switch settings.Storage.Type {
	case "redis":
		return redis.MustNew(settings.Storage.RedisHost)
	default:
		return memory.MustNew(true, path)
	}
}

func MustNewCombatContainer(ctx context.Context) Container {
	level, err := zerolog.ParseLevel(shared.AppSettings.LogLevel)
	utils.Check(err)

	zerolog.SetGlobalLevel(level)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	persistentStorage := makeStorage(&shared.AppSettings)

	kandinskySupplier, err := kandinsky_supplier.New(
		context.Background(),
		shared.AppSettings.KandinskyKey,
		shared.AppSettings.KandinskySecret,
	)
	if err != nil {
		log.Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.create.kandinsky.supplier").Send()
	}

	locator := ip_locator.Supplier{}

	dbRepository, err := db_repository.New(shared.AppSettings.Storage.MongoDbUrl)
	utils.Check(err)

	analiticsService, err := analitics.New(dbRepository)
	utils.Check(err)

	telegramPresentation := telegram.MustNewTelegramPresentation(
		ctx,
		persistentStorage,
		kandinskySupplier,
		&locator,
		dbRepository,
		analiticsService,
	)

	jobService, err := job.New(ctx, dbRepository)
	utils.Check(err)

	container := Container{telegramPresentation, jobService}

	return container
}
