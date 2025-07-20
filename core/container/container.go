package container

import (
	"context"

	"github.com/glebarez/sqlite"
	"github.com/teadove/fun_telegram/core/repository/db_repository"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

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

	db, err := gorm.Open(sqlite.Open(shared.AppSettings.SQLiteFile), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy:         schema.NamingStrategy{SingularTable: true},
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
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

	telegramPresentation := telegram.MustNewTelegramPresentation(
		ctx,
		protoClient,
		persistentStorage,
		kandinskySupplier,
		&locator,
		analiticsService,
		resourceService,
		dbRepository,
	)

	container := Container{telegramPresentation, jobService}

	return container
}
