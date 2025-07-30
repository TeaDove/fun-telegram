package container

import (
	"context"

	"fun_telegram/core/repository/db_repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"fun_telegram/core/supplier/ds_supplier"

	"fun_telegram/core/presentation/telegram"
	"fun_telegram/core/service/analitics"
	"fun_telegram/core/shared"

	"github.com/pkg/errors"
)

type Container struct {
	Presentation *telegram.Presentation
}

func MustNewCombatContainer(ctx context.Context) Container {
	dsSupplier := ds_supplier.New()

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

	telegramPresentation := telegram.MustNewTelegramPresentation(protoClient, analiticsService, dbRepository)

	container := Container{telegramPresentation}

	return container
}
