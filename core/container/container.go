package container

import (
	"context"

	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/presentation/telegram"
	"github.com/teadove/fun_telegram/core/repository/ch_repository"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"github.com/teadove/fun_telegram/core/repository/redis_repository"
	"github.com/teadove/fun_telegram/core/service/analitics"
	"github.com/teadove/fun_telegram/core/service/job"
	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/shared"
	"github.com/teadove/fun_telegram/core/supplier/ip_locator"
	"github.com/teadove/fun_telegram/core/supplier/kandinsky_supplier"
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
		zerolog.Ctx(ctx).
			Error().
			Stack().
			Err(errors.WithStack(err)).
			Str("status", "failed.to.create.kandinsky.supplier").
			Send()
	}

	locator := ip_locator.Supplier{}

	dbRepository, err := mongo_repository.New()
	shared.Check(ctx, err)

	chRepository, err := ch_repository.New(ctx)
	shared.Check(ctx, err)

	dsSupplier, err := ds_supplier.New(ctx)
	shared.Check(ctx, err)

	resourceService, err := resource.New(ctx)
	shared.Check(ctx, err)

	analiticsService, err := analitics.New(dbRepository, chRepository, dsSupplier, resourceService)
	shared.Check(ctx, err)

	protoClient := telegram.MustNewProtoClient(ctx)

	jobService, err := job.New(ctx, dbRepository, chRepository, map[string]job.ServiceChecker{
		"MongoDB":    {Checker: dbRepository.Ping, ForFrequent: true},
		"Telegram":   {Checker: protoClient.Ping, ForFrequent: true},
		"Redis":      {Checker: persistentStorage.Ping, ForFrequent: true},
		"ClickHouse": {Checker: chRepository.Ping, ForFrequent: true},
		"Kandinsky":  {Checker: kandinskySupplier.Ping},
		"IpLocator":  {Checker: locator.Ping},
		"DSSupplier": {Checker: dsSupplier.Ping},
	})
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
