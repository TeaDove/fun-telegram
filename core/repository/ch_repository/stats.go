package ch_repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/schemas"
)

func (r *Repository) StatsForDatabase(
	ctx context.Context,
) (map[string]schemas.StorageStats, error) {
	rows, err := r.conn.Query(ctx, `
SELECT table as table,
       sum(bytes) AS bytes,
       sum(rows)  AS count
FROM system.parts
WHERE active and database = ?
GROUP BY 1
ORDER BY bytes DESC
`, r.databaseName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get stats from database")
	}

	type record struct {
		Table          string `ch:"table"`
		TotalSizeBytes uint64 `ch:"bytes"`
		Count          uint64 `ch:"count"`
	}

	map_ := make(map[string]schemas.StorageStats, 3)

	for rows.Next() {
		var row record
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		map_[row.Table] = schemas.StorageStats{
			TotalSizeBytes:           int(row.TotalSizeBytes),
			Count:                    int(row.Count),
			AvgObjWithIndexSizeBytes: int(row.TotalSizeBytes / row.Count),
		}
	}

	return map_, nil
}
