package job

import (
	"context"
	"github.com/go-co-op/gocron"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/repository/ch_repository"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
	"github.com/teadove/goteleout/internal/schemas"
	"github.com/teadove/goteleout/internal/shared"
	"time"
)

type Service struct {
	mongoRepository *mongo_repository.Repository
	chRepository    *ch_repository.Repository

	checkers map[string]ServiceChecker
}

func New(
	ctx context.Context,
	dbRepository *mongo_repository.Repository,
	chRepository *ch_repository.Repository,
	checkers map[string]ServiceChecker,
) (*Service, error) {
	ctx = shared.AddModuleCtx(ctx, "job")
	r := Service{mongoRepository: dbRepository, checkers: checkers, chRepository: chRepository}

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

func (r *Service) Stats(ctx context.Context) (map[string]map[string]schemas.StorageStats, error) {
	statsByDatabase := make(map[string]map[string]schemas.StorageStats, 2)

	stats, err := r.mongoRepository.StatsForDatabase(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	statsByDatabase["MongoDB"] = stats

	stats, err = r.chRepository.StatsForDatabase(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	statsByDatabase["Clickhouse"] = stats

	return statsByDatabase, nil
}

func (r *Service) deleteOldMessagesChecked(ctx context.Context) {
	_, err := r.DeleteOldMessages(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.delete.old.messages").Send()
	}
}

func (r *Service) DeleteOldMessages(ctx context.Context) (DeleteOldMessagesOutput, error) {
	stats, err := r.mongoRepository.StatsForMessages(ctx)
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

	_, err = r.mongoRepository.DeleteMessagesOldWithCount(ctx, int64(countToDelete))
	if err != nil {
		return DeleteOldMessagesOutput{}, errors.WithStack(err)
	}

	bytesFreed, err := r.mongoRepository.ReleaseMemory(ctx)
	if err != nil {
		return DeleteOldMessagesOutput{}, errors.WithStack(err)
	}

	newStats, err := r.mongoRepository.StatsForMessages(ctx)
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
