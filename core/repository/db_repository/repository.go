package db_repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
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
	err := r.db.
		WithContext(ctx).
		AutoMigrate(
			&Message{},
			&Member{},
			&Chat{},
			&User{},
		)
	if err != nil {
		return errors.Wrap(err, "failed to migrate database")
	}

	zerolog.Ctx(ctx).Info().Msg("pg.migrated.successfully")

	return nil
}
