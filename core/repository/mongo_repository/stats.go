package mongo_repository

import (
	"context"

	"github.com/kamva/mgm/v3"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/schemas"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *Repository) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) StatsForMessages(ctx context.Context) (schemas.StorageStats, error) {
	return r.StatsForTable(ctx, mgm.CollName(&Message{}))
}

func (r *Repository) StatsForTable(ctx context.Context, collName string) (schemas.StorageStats, error) {
	result := r.client.Database(databaseName).RunCommand(ctx, bson.M{"collStats": collName})

	var document bson.M
	err := result.Decode(&document)
	if err != nil {
		return schemas.StorageStats{}, errors.WithStack(err)
	}

	stats := schemas.StorageStats{}

	count, ok := document["count"].(int32)
	if !ok {
		return schemas.StorageStats{}, errors.New("failed to get count from stats")
	}

	stats.Count = int(count)

	totalSize, ok := document["totalSize"].(int32)
	if !ok {
		return schemas.StorageStats{}, errors.New("failed to get totalSize from stats")
	}

	stats.TotalSizeBytes = int(totalSize)

	if stats.Count != 0 {
		stats.AvgObjWithIndexSizeBytes = stats.TotalSizeBytes / stats.Count
	}

	return stats, nil
}

func (r *Repository) StatsForDatabase(ctx context.Context) (map[string]schemas.StorageStats, error) {
	colls, err := r.client.Database(databaseName).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	map_ := make(map[string]schemas.StorageStats, len(colls))
	for _, coll := range colls {
		map_[coll], err = r.StatsForTable(ctx, coll)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return map_, nil
}

func (r *Repository) ReleaseMemory(ctx context.Context) (int, error) {
	bytesFreed := 0
	colls, err := r.client.Database(databaseName).ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return 0, errors.WithStack(err)
	}

	for _, coll := range colls {
		result := r.client.Database(databaseName).RunCommand(ctx, bson.M{"compact": coll})

		var document bson.M

		err = result.Decode(&document)
		if err != nil {
			return 0, errors.WithStack(err)
		}

		bytesFreedColl, ok := document["bytesFreed"].(int32)
		if !ok {
			return 0, errors.New("failed to get bytesFreed")
		}

		bytesFreed += int(bytesFreedColl)
	}

	return bytesFreed, nil
}
