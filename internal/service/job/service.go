package job

import (
	"context"
	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/utils"
	"time"
)

type Service struct {
	dbRepository *db_repository.Repository

	// nolint: containedctx
	ctx      context.Context
	checkers map[string]ServiceChecker
}

func New(ctx context.Context, dbRepository *db_repository.Repository, checkers map[string]ServiceChecker) (*Service, error) {
	r := Service{dbRepository: dbRepository, ctx: utils.AddModuleCtx(ctx, "job"), checkers: checkers}

	scheduler := gocron.NewScheduler(time.UTC)

	tomorrow := time.Now().UTC()
	tomorrowNight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day()+1, 0, 0, 0, 0, tomorrow.Location())

	_, err := scheduler.
		Every(24 * time.Hour).StartAt(tomorrowNight).
		Do(r.DeleteOldMessages)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create scheduler")
	}

	scheduler.StartAsync()

	return &r, nil
}

func (r *Service) DeleteOldMessages() {
	count, err := r.dbRepository.MessageDeleteOld(r.ctx)
	if err != nil {
		zerolog.Ctx(r.ctx).
			Error().
			Stack().
			Err(errors.WithStack(err)).
			Str("status", "failed.to.delete.old.messages").
			Send()
	}

	zerolog.Ctx(r.ctx).Info().Str("status", "old.messages.deleted").Int64("count", count).Send()
}
