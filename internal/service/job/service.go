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
	ctx          context.Context
}

func New(dbRepository *db_repository.Repository) (*Service, error) {
	r := Service{dbRepository: dbRepository, ctx: utils.GetModuleCtx("job")}

	scheduler := gocron.NewScheduler(time.UTC)

	_, err := scheduler.
		Every(24 * time.Hour).
		Do(r.DeleteOldMessages)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create scheduler")
	}

	scheduler.StartAsync()

	return &r, nil
}

func (r *Service) DeleteOldMessages() {
	err := r.dbRepository.MessageDeleteOld(r.ctx)
	if err != nil {
		zerolog.Ctx(r.ctx).Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.delete.old.messages").Send()
	}
}
