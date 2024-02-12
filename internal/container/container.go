package container

import (
	"context"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
)

type Container struct {
	Presentation *telegram.Presentation
	JobService   *job.Service
}

func MustNewCombatContainer(ctx context.Context) Container {
	persistentStorage := redis_repository.MustNew()

	kandinskySupplier, err := kandinsky_supplier.New(
		ctx,
		shared.AppSettings.KandinskyKey,
		shared.AppSettings.KandinskySecret,
	)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.create.kandinsky.supplier").Send()
	}

	locator := ip_locator.Supplier{}

	dbRepository, err := mongo_repository.New()
	shared.Check(ctx, err)

	chRepository, err := ch_repository.New(ctx)
	shared.Check(ctx, err)

	analiticsService, err := analitics.New(dbRepository, chRepository)
	shared.Check(ctx, err)

	protoClient := telegram.MustNewProtoClient(ctx)

	jobService, err := job.New(ctx, dbRepository, map[string]job.ServiceChecker{
		"MongoDB":    {Checker: dbRepository.Ping, ForFrequent: true},
		"Telegram":   {Checker: protoClient.Ping, ForFrequent: true},
		"Redis":      {Checker: persistentStorage.Ping, ForFrequent: true},
		"ClickHouse": {Checker: chRepository.Ping, ForFrequent: true},
		"Kandinsky":  {Checker: kandinskySupplier.Ping},
		"IpLocator":  {Checker: locator.Ping},
	})
	shared.Check(ctx, err)

	resourceService, err := resource.New(ctx)
	shared.Check(ctx, err)

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
