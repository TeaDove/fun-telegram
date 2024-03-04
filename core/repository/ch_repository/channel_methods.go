package ch_repository

import (
	"context"

	"github.com/pkg/errors"
)

func (r *Repository) ChannelInsert(ctx context.Context, channel *Channel) error {
	err := r.conn.AsyncInsert(ctx, `
INSERT INTO channel VALUES (
			?, ?, ?, ?, ?, ?
		)`,
		true,
		channel.TgId,
		channel.TgTitle,
		channel.TgUsername,
		channel.UploadedAt,
		channel.ParticipantCount,
		channel.RecommendationsIds,
	)
	if err != nil {
		return errors.Wrap(err, "failed to async insert")
	}

	return nil
}

func (r *Repository) ChannelEdgeBatchInsert(ctx context.Context, channels []ChannelEdge) error {
	batch, err := r.conn.PrepareBatch(ctx, `INSERT INTO channel_edge`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare batch")
	}

	for _, channel := range channels {
		err = batch.Append(
			channel.TgIdIn,
			channel.TgIdOut,
			channel.Order,
		)
		if err != nil {
			return errors.Wrap(err, "failed to append to batch")
		}
	}

	err = batch.Send()
	if err != nil {
		return errors.Wrap(err, "failed to batch send")
	}

	return nil
}

func (r *Repository) ChannelBatchInsert(ctx context.Context, channels []Channel) error {
	batch, err := r.conn.PrepareBatch(ctx, `INSERT INTO channel`)
	if err != nil {
		return errors.Wrap(err, "failed to prepare batch")
	}

	for _, channel := range channels {
		err = batch.Append(
			channel.TgId,
			channel.TgTitle,
			channel.TgUsername,
			channel.UploadedAt,
			channel.ParticipantCount,
			channel.RecommendationsIds,
		)
		if err != nil {
			return errors.Wrap(err, "failed to append to batch")
		}
	}

	err = batch.Send()
	if err != nil {
		return errors.Wrap(err, "failed to batch send")
	}

	return nil
}

func (r *Repository) ChannelSelect(ctx context.Context, id int64) (Channel, error) {
	row := r.conn.QueryRow(ctx, `
select tg_id, tg_title, tg_username, uploaded_at, participant_count, recommendations_ids from channel final
	where tg_id = ? 
`, id)
	if row.Err() != nil {
		return Channel{}, errors.Wrap(row.Err(), "failed to select row from clickhouse")
	}

	var channel Channel
	err := row.ScanStruct(&channel)
	if err != nil {
		return Channel{}, errors.Wrap(err, "failed to scan row")
	}

	return channel, nil
}

func (r *Repository) ChannelSelectById(ctx context.Context, id []int64) (Channels, error) {
	rows, err := r.conn.Query(ctx, `
select tg_id, tg_title, tg_username, uploaded_at, participant_count, recommendations_ids from channel final
	where tg_id in ? 
`, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select rows from clickhouse")
	}

	output := make(Channels, 0, len(id))
	for rows.Next() {
		row := Channel{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, nil
}
