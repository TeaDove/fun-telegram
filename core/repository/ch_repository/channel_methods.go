package ch_repository

import (
	"context"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/pkg/errors"
)

func (r *Repository) ChannelInsert(ctx context.Context, channel *Channel) error {
	err := r.conn.AsyncInsert(ctx, `
INSERT INTO channel VALUES (
			?, ?, ?, ?, ?, ?, ?, ?
		)`,
		true,
		channel.TgId,
		channel.TgTitle,
		channel.TgUsername,
		channel.UploadedAt,
		channel.ParticipantCount,
		channel.RecommendationsIds,
		channel.IsLeaf,
		channel.TgAbout,
	)
	if err != nil {
		return errors.Wrap(err, "failed to async insert")
	}

	return nil
}

func (r *Repository) ChannelEdgeBatchInsert(ctx context.Context, channels ChannelsEdges) error {
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
			channel.IsLeaf,
			channel.TgAbout,
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

func (r *Repository) ChannelSelectByUsername(
	ctx context.Context,
	username string,
) (Channel, error) {
	row := r.conn.QueryRow(ctx, `
select tg_id, tg_title, tg_username, uploaded_at, participant_count, recommendations_ids, is_leaf, tg_about from channel final
	where tg_username = ? 
`, username)
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

func (r *Repository) ChannelSelectById(ctx context.Context, id int64) (Channel, error) {
	row := r.conn.QueryRow(ctx, `
select tg_id, tg_title, tg_username, uploaded_at, participant_count, recommendations_ids, is_leaf, tg_about from channel final
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

func (r *Repository) ChannelSelectByIds(ctx context.Context, id []int64) (Channels, error) {
	rows, err := r.conn.Query(ctx, `
select tg_id, tg_title, tg_username, uploaded_at, participant_count, recommendations_ids, is_leaf, tg_about from channel final
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

func (r *Repository) ChannelSelect(ctx context.Context) (Channels, error) {
	rows, err := r.conn.Query(ctx, `
select tg_id, tg_title, tg_username, uploaded_at, participant_count, recommendations_ids, is_leaf, tg_about 
	from channel final
	order by tg_id 
`)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select rows from clickhouse")
	}

	output := make(Channels, 0, 100)

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

func (r *Repository) ChannelEdgesSelect(
	ctx context.Context,
	maxOrder int64,
) (ChannelsEdges, error) {
	rows, err := r.conn.Query(ctx, `
select tg_id_in, tg_id_out, order from channel_edge final order by tg_id_in, tg_id_out and order <= ?
`, maxOrder)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select rows from clickhouse")
	}

	output := make([]ChannelEdge, 0, 100)

	for rows.Next() {
		row := ChannelEdge{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, nil
}

func (r *Repository) ChannelEdgesSelectById(
	ctx context.Context,
	tgIdIn []int64,
	maxOrder int64,
) (ChannelsEdges, error) {
	rows, err := r.conn.Query(ctx, `
	select tg_id_in, tg_id_out, order from channel_edge final where tg_id_in in ? and order <= ?   
`, tgIdIn, maxOrder)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select rows from clickhouse")
	}

	output := make(ChannelsEdges, 0, 100)

	for rows.Next() {
		row := ChannelEdge{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, nil
}

func (r *Repository) ChannelEdgesSelectDFS(
	ctx context.Context,
	tgUsername string,
	depth int64,
	maxOrder int64,
) (ChannelsEdges, error) {
	channel, err := r.ChannelSelectByUsername(ctx, tgUsername)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select channel by username")
	}

	channelEdgesResult := mapset.NewSet[ChannelEdge]()
	tgIds := []int64{channel.TgId}

	for _ = range depth {
		channelEdges, err := r.ChannelEdgesSelectById(ctx, tgIds, maxOrder)
		if err != nil {
			return nil, errors.Wrap(err, "failed to select channel edges by id")
		}

		channelEdgesResult.Append(channelEdges...)
		tgIds = channelEdges.ToOutIds()
	}

	return channelEdgesResult.ToSlice(), nil
}
