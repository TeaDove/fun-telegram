package container

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/presentation/telegram"
	"github.com/teadove/goteleout/internal/repository/ch_repository"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"github.com/teadove/goteleout/internal/repository/redis_repository"
	"github.com/teadove/goteleout/internal/service/analitics"
	"github.com/teadove/goteleout/internal/service/job"
	"github.com/teadove/goteleout/internal/service/resource"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/supplier/ip_locator"
	"github.com/teadove/goteleout/internal/supplier/kandinsky_supplier"
	"github.com/teadove/goteleout/internal/utils"
	"os"
)

type Container struct {
	Presentation *telegram.Presentation
	JobService   *job.Service
}

func MustNewCombatContainer(ctx context.Context) Container {
	level, err := zerolog.ParseLevel(shared.AppSettings.LogLevel)
	utils.Check(err)

	zerolog.SetGlobalLevel(level)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	persistentStorage := redis_repository.MustNew()

	kandinskySupplier, err := kandinsky_supplier.New(
		ctx,
		shared.AppSettings.KandinskyKey,
		shared.AppSettings.KandinskySecret,
	)
	if err != nil {
		log.Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.create.kandinsky.supplier").Send()
	}

	locator := ip_locator.Supplier{}

	dbRepository, err := mongo_repository.New()
	utils.Check(err)

	analiticsService, err := analitics.New(dbRepository)
	utils.Check(err)

	protoClient := telegram.MustNewProtoClient(ctx)

	chRepository, err := ch_repository.New(ctx)
	utils.Check(err)

	jobService, err := job.New(ctx, dbRepository, map[string]job.ServiceChecker{
		"MongoDB":    {Checker: dbRepository.Ping, ForFrequent: true},
		"Telegram":   {Checker: protoClient.Ping, ForFrequent: true},
		"Redis":      {Checker: persistentStorage.Ping, ForFrequent: true},
		"ClickHouse": {Checker: chRepository.Ping, ForFrequent: true},
		"Kandinsky":  {Checker: kandinskySupplier.Ping},
		"IpLocator":  {Checker: locator.Ping},
	})
	utils.Check(err)

	resourceService, err := resource.New(ctx)
	utils.Check(err)

	telegramPresentation := telegram.MustNewTelegramPresentation(
		ctx,
		protoClient,
		persistentStorage,
		kandinskySupplier,
		&locator,
		dbRepository,
		analiticsService,
		jobService,
		resourceService,
	)

	container := Container{telegramPresentation, jobService}

	return container
}
