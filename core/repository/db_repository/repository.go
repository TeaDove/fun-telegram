package db_repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(ctx context.Context, db *gorm.DB) (*Repository, error) {
	r := Repository{db: db}

	err := r.Migrate(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to migrate repository")
	}

	return &r, nil
}

func (r *Repository) Migrate(ctx context.Context) error {
	oldLogger := r.db.Logger
	r.db.Logger = oldLogger.LogMode(logger.Silent)
	defer func() { r.db.Logger = oldLogger }()

	err := r.db.
		WithContext(ctx).
		AutoMigrate(
			&Message{},
			&Member{},
			&Channel{},
			&ChannelEdge{},
			&Chat{},
			&User{},
			&RestartMessage{},
			&PingMessage{},
			&Image{},
			&KandinskyImage{},
			&TgImage{},
		)
	if err != nil {
		return errors.Wrap(err, "failed to migrate database")
	}

	zerolog.Ctx(ctx).Info().Msg("pg.migrated.successfully")
	return nil
}

func (r *Repository) Ping(ctx context.Context) error {
	var v int64

	err := r.db.WithContext(ctx).Raw("select 1").Scan(&v).Error
	if err != nil {
		return errors.Wrap(err, "failed to ping database")
	}

	if v == 0 {
		return errors.New("non one response")
	}

	return nil
}
