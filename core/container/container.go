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
	"github.com/teadove/fun_telegram/core/presentation/telegram"
	"github.com/teadove/fun_telegram/core/service/analitics"
	"github.com/teadove/fun_telegram/core/shared"
)

type Container struct {
	Presentation *telegram.Presentation
}

func MustNewCombatContainer(ctx context.Context) Container {
	dsSupplier, err := ds_supplier.New(ctx)
	shared.Check(ctx, err)

	db, err := gorm.Open(sqlite.Open(shared.AppSettings.SQLiteFile), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy:         schema.NamingStrategy{SingularTable: true},
		Logger:                 logger.Default.LogMode(logger.Silent),
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		shared.FancyPanic(ctx, errors.Wrap(err, "failed to init pg client"))
	}

	dbRepository, err := db_repository.NewRepository(ctx, db)
	if err != nil {
		shared.FancyPanic(ctx, errors.Wrap(err, "failed to init pg repository"))
	}

	analiticsService, err := analitics.New(dsSupplier, dbRepository)
	shared.Check(ctx, err)

	protoClient, err := telegram.NewProtoClient(ctx)
	shared.Check(ctx, err)

	telegramPresentation := telegram.MustNewTelegramPresentation(
		ctx,
		protoClient,
		analiticsService,
		dbRepository,
	)

	container := Container{telegramPresentation}

	return container
}
