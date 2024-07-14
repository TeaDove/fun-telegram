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
			// TODO fix Channel
			//&Channel{},
			&ChannelEdge{},
			&Chat{},
			&User{},
			&RestartMessage{},
			&PingMessage{},
			&Image{},
		)
	if err != nil {
		return errors.Wrap(err, "failed to migrate database")
	}

	zerolog.Ctx(ctx).Info().Msg("pg.migrated.successfully")
	return nil
}

func (r *Repository) Ping(ctx context.Context) error {
	var row struct {
		v int
	}

	err := r.db.WithContext(ctx).Raw("select 1").Scan(&row).Error
	if err != nil {
		return errors.Wrap(err, "failed to ping database")
	}

	if row.v == 0 {
		return errors.New("non one response")
	}

	return nil
}
