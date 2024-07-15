package container

import (
	"context"

	"github.com/teadove/fun_telegram/core/infrastructure/pg"
	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/presentation/telegram"
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

	dsSupplier, err := ds_supplier.New(ctx)
	shared.Check(ctx, err)

	resourceService, err := resource.New(ctx)
	shared.Check(ctx, err)

	db, err := pg.NewClientFromSettings()
	if err != nil {
		shared.FancyPanic(ctx, errors.Wrap(err, "failed to init pg client"))
	}

	dbRepository, err := db_repository.NewRepository(ctx, db)
	if err != nil {
		shared.FancyPanic(ctx, errors.Wrap(err, "failed to init pg repository"))
	}

	analiticsService, err := analitics.New(
		dsSupplier,
		resourceService,
		dbRepository,
	)
	shared.Check(ctx, err)

	protoClient, err := telegram.NewProtoClient(ctx)
	shared.Check(ctx, err)

	jobService, err := job.New(ctx, map[string]job.ServiceChecker{
		"Telegram":   {Checker: protoClient.Ping, ForFrequent: true},
		"Redis":      {Checker: persistentStorage.Ping, ForFrequent: true},
		"Postgres":   {Checker: dbRepository.Ping, ForFrequent: true},
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
		analiticsService,
		jobService,
		resourceService,
		dbRepository,
	)

	container := Container{telegramPresentation, jobService}

	return container
}
