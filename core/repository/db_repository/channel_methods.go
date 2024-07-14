package db_repository

import (
	"context"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/pkg/errors"
)

func (r *Repository) ChannelInsert(ctx context.Context, channel *Channel) error {
	err := r.db.WithContext(ctx).Create(&channel).Error
	if err != nil {
		return errors.Wrap(err, "failed to insert channel")
	}

	return nil
}

func (r *Repository) ChannelEdgeBatchInsert(ctx context.Context, channels ChannelsEdges) error {
	err := r.db.WithContext(ctx).Create(&channels).Error
	if err != nil {
		return errors.Wrap(err, "failed to insert channels edges")
	}

	return nil
}

func (r *Repository) ChannelBatchInsert(ctx context.Context, channels []Channel) error {
	err := r.db.WithContext(ctx).Create(&channels).Error
	if err != nil {
		return errors.Wrap(err, "failed to insert channels")
	}

	return nil
}

func (r *Repository) ChannelSelectByUsername(
	ctx context.Context,
	tgUsername string,
) (Channel, error) {
	var channel Channel

	err := r.db.WithContext(ctx).First(&channel, "tg_username = ?", tgUsername).Error
	if err != nil {
		return Channel{}, errors.Wrap(err, "failed to query channel by username")
	}

	return channel, nil
}

func (r *Repository) ChannelSelectById(ctx context.Context, tgId int64) (Channel, error) {
	var channel Channel

	err := r.db.WithContext(ctx).First(&channel, "tg_id = ?", tgId).Error
	if err != nil {
		return Channel{}, errors.Wrap(err, "failed to query channel by id")
	}

	return channel, nil
}

func (r *Repository) ChannelSelectByIds(ctx context.Context, tgIds []int64) (Channels, error) {
	var channels []Channel

	err := r.db.WithContext(ctx).Find(&channels, "tg_id in (?)", tgIds).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to query channel by ids")
	}

	return channels, nil
}

func (r *Repository) ChannelEdgesSelect(
	ctx context.Context,
	maxOrder int64,
) (ChannelsEdges, error) {
	var out ChannelsEdges
	err := r.db.WithContext(ctx).Find(out, "order <= ?", maxOrder).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to query channel edges")
	}

	return out, nil
}

func (r *Repository) ChannelEdgesSelectById(
	ctx context.Context,
	tgIdIn []int64,
	maxOrder int64,
) (ChannelsEdges, error) {
	var out ChannelsEdges
	err := r.db.WithContext(ctx).Find(out, "order <= ? AND tg_id_in = ?", maxOrder, tgIdIn).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to query channel edges")
	}

	return out, nil
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

	for range depth {
		channelEdges, err := r.ChannelEdgesSelectById(ctx, tgIds, maxOrder)
		if err != nil {
			return nil, errors.Wrap(err, "failed to select channel edges by id")
		}

		channelEdgesResult.Append(channelEdges...)
		tgIds = channelEdges.ToOutIds()
	}

	return channelEdgesResult.ToSlice(), nil
}
