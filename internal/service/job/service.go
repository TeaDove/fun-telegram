package job

import (
	"context"
	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/utils"
	"time"
)

type Service struct {
	dbRepository *db_repository.Repository

	checkers map[string]ServiceChecker
}

func New(ctx context.Context, dbRepository *db_repository.Repository, checkers map[string]ServiceChecker) (*Service, error) {
	ctx = utils.AddModuleCtx(ctx, "job")
	r := Service{dbRepository: dbRepository, checkers: checkers}

	scheduler := gocron.NewScheduler(time.UTC)

	tomorrow := time.Now().UTC()
	tomorrowNight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day()+1, 0, 0, 0, 0, tomorrow.Location())

	_, err := scheduler.
		Every(24*time.Hour).StartAt(tomorrowNight).
		Do(r.deleteOldMessagesChecked, ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create scheduler")
	}

	scheduler.StartAsync()

	return &r, nil
}

type DeleteOldMessagesOutput struct {
	OldCount   int
	NewCount   int
	OldSize    int
	NewSize    int
	BytesFreed int
}

func (r *Service) deleteOldMessagesChecked(ctx context.Context) {
	_, err := r.DeleteOldMessages(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.delete.old.messages").Send()
	}
}

func (r *Service) DeleteOldMessages(ctx context.Context) (DeleteOldMessagesOutput, error) {
	stats, err := r.dbRepository.StatsForMessages(ctx)
	if err != nil {
		return DeleteOldMessagesOutput{}, errors.WithStack(err)
	}

	desiredSizeInBytes := 1024 * 1024 * shared.AppSettings.MessagesMaxSizeMB
	if desiredSizeInBytes > stats.TotalSizeBytes {
		return DeleteOldMessagesOutput{}, nil
	}

	sizeToDelete := stats.TotalSizeBytes - desiredSizeInBytes

	countToDelete := sizeToDelete / stats.AvgObjWithIndexSizeBytes
	if countToDelete == 0 {
		return DeleteOldMessagesOutput{}, nil
	}

	_, err = r.dbRepository.DeleteMessagesOldWithCount(ctx, int64(countToDelete))
	if err != nil {
		return DeleteOldMessagesOutput{}, errors.WithStack(err)
	}

	bytesFreed, err := r.dbRepository.ReleaseMemory(ctx)
	if err != nil {
		return DeleteOldMessagesOutput{}, errors.WithStack(err)
	}

	newStats, err := r.dbRepository.StatsForMessages(ctx)
	if err != nil {
		return DeleteOldMessagesOutput{}, errors.WithStack(err)
	}

	output := DeleteOldMessagesOutput{
		OldCount:   stats.Count,
		NewCount:   newStats.Count,
		OldSize:    stats.TotalSizeBytes,
		NewSize:    newStats.TotalSizeBytes,
		BytesFreed: bytesFreed,
	}
	zerolog.Ctx(ctx).
		Info().
		Str("status", "old.messages.deleted").
		Interface("output", output).
		Send()

	return output, nil
}
