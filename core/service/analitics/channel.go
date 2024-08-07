package analitics

import (
	"context"
	"runtime"
	"time"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/shared"
	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func (r *Service) ChannelInsert(ctx context.Context, channel *db_repository.Channel) error {
	channel.CreatedInDBAt = time.Now().UTC()

	err := r.dbRepository.ChannelInsert(ctx, channel)
	if err != nil {
		return errors.Wrap(err, "failed to insert channel")
	}

	zerolog.Ctx(ctx).
		Debug().
		Str("status", "channel.inserted").
		Interface("title", channel.TgTitle).
		Send()

	return nil
}

func (r *Service) ChannelSelect(ctx context.Context, id int64) (db_repository.Channel, error) {
	channel, err := r.dbRepository.ChannelSelectById(ctx, id)
	if err != nil {
		return db_repository.Channel{}, errors.Wrap(err, "failed to select channel")
	}

	return channel, nil
}

var channelDataTtl = time.Hour * 24 * 60

func (r *Service) channelEdgeBatchInsert(
	ctx context.Context,
	channels db_repository.Channels,
) error {
	edges := make([]db_repository.ChannelEdge, 0, len(channels)*2)

	for _, channelIn := range channels {
		for idx, channelOut := range channelIn.RecommendationsIds {
			edges = append(edges, db_repository.ChannelEdge{
				TgIdIn:  channelIn.TgId,
				TgIdOut: channelOut,
				Order:   int64(idx),
			})
		}
	}

	err := r.dbRepository.ChannelEdgeBatchInsert(ctx, edges)
	if err != nil {
		return errors.Wrap(err, "failed to batch insert channel edges")
	}

	return nil
}

// ChannelBatchInsert
// nolint: cyclop
// TODO fix cyclop
func (r *Service) ChannelBatchInsert(ctx context.Context, channels []db_repository.Channel) error {
	channelIds := make([]int64, len(channels))
	for idx, channel := range channels {
		channelIds[idx] = channel.TgId
	}

	oldChannels, err := r.dbRepository.ChannelSelectByIds(ctx, channelIds)
	if err != nil {
		return errors.Wrap(err, "failed to select channels")
	}

	oldChannelsMap := oldChannels.ToMap()

	leafsCount := 0
	alreadyProcessed := 0
	newChannels := make([]db_repository.Channel, 0, len(channels))

	for _, channel := range channels {
		oldChannel, ok := oldChannelsMap[channel.TgId]
		if ok && time.Since(oldChannel.CreatedInDBAt) < channelDataTtl && !oldChannel.IsLeaf &&
			channel.IsLeaf {
			alreadyProcessed++
			continue
		}

		channel.CreatedInDBAt = time.Now().UTC()
		if channel.RecommendationsIds == nil {
			leafsCount++
		}

		newChannels = append(newChannels, channel)
	}

	err = r.dbRepository.ChannelBatchInsert(ctx, newChannels)
	if err != nil {
		return errors.Wrap(err, "failed to insert channels")
	}

	err = r.channelEdgeBatchInsert(ctx, newChannels)
	if err != nil {
		return errors.Wrap(err, "failed to insert channel edges")
	}

	zerolog.Ctx(ctx).Debug().
		Str("status", "channels.inserted").
		Int("count", len(newChannels)).
		Int("leafs", leafsCount).
		Int("non_leafs", len(newChannels)-leafsCount).
		Int("already_processed", alreadyProcessed).
		Send()

	return nil
}

// DumpChannels
// nolint: cyclop
// TODO fix self-loops
func (r *Service) DumpChannels(
	ctx context.Context,
	username string,
	depth int64,
	maxOrder int64,
) ([]File, error) {
	files := make([]File, 0, 3)

	var (
		err          error
		channelEdges db_repository.ChannelsEdges
	)

	if username != "" {
		channelEdges, err = r.dbRepository.ChannelEdgesSelectDFS(ctx, username, depth, maxOrder)
		if err != nil {
			return nil, errors.Wrap(err, "failed to dsf channel edges")
		}
	} else {
		channelEdges, err = r.dbRepository.ChannelEdgesSelect(ctx, maxOrder)
		if err != nil {
			return nil, errors.Wrap(err, "failed to dsf channel edges")
		}
	}

	uniqueIds := channelEdges.ToIds()

	if len(uniqueIds) == 0 {
		return nil, errors.New("no channels found")
	}

	zerolog.Ctx(ctx).
		Info().
		Str("status", "dumping.channels").
		Int("edges.count", len(channelEdges)).
		Int("channels.count", len(uniqueIds)).
		Send()

	file, err := r.dumpChannelsEdgeParquet(channelEdges)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dump channels edges")
	}

	err = file.Compress()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compress file: %s", file.Filename())
	}

	files = append(files, file)

	file, err = r.dumpChannelsParquet(ctx, uniqueIds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dump channels")
	}

	err = file.Compress()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compress file: %s", file.Filename())
	}

	files = append(files, file)

	file, err = r.dumpMessagesParquet(ctx, uniqueIds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dump messages")
	}

	err = file.Compress()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to compress file: %s", file.Filename())
	}

	files = append(files, file)

	runtime.GC()

	return files, nil
}

type AnaliseChannelInput struct {
	TgUsername string
	Depth      int64
	MaxOrder   int64

	Locale resource.Locale
}

func (r *Service) AnaliseChannel(ctx context.Context, input *AnaliseChannelInput) (File, error) {
	rootChannel, err := r.dbRepository.ChannelSelectByUsername(ctx, input.TgUsername)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to select channel by username")
	}

	channelEdges, err := r.dbRepository.ChannelEdgesSelectDFS(
		ctx,
		input.TgUsername,
		input.Depth,
		input.MaxOrder,
	)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to dsf channel edges")
	}

	if len(channelEdges) == 0 {
		return File{}, errors.New("no channel edges found")
	}

	channels, err := r.dbRepository.ChannelSelectByIds(ctx, channelEdges.ToIds())
	if err != nil {
		return File{}, errors.Wrap(err, "failed to select channel")
	}

	zerolog.Ctx(ctx).Info().
		Str("status", "analysing.channel.begin").
		Int("edges", len(channelEdges)).
		Int("channels", len(channels)).
		Str("title", rootChannel.TgTitle).
		Send()

	channelsMap := channels.ToMap()

	edges := make([]ds_supplier.GraphEdge, 0, len(channelEdges))
	for _, edge := range channelEdges {
		edges = append(edges, ds_supplier.GraphEdge{
			First:  shared.ReplaceNonAsciiWithSpace(channelsMap[edge.TgIdIn].TgTitle),
			Second: shared.ReplaceNonAsciiWithSpace(channelsMap[edge.TgIdOut].TgTitle),
			Weight: float64(input.MaxOrder - edge.Order),
		})
	}

	nodes := make(map[string]ds_supplier.GraphNode, len(channels))
	for _, node := range channels {
		nodes[shared.ReplaceNonAsciiWithSpace(node.TgTitle)] = ds_supplier.GraphNode{
			Weight: float64(node.ParticipantCount),
		}
	}

	zerolog.Ctx(ctx).
		Debug().
		Str("status", "channel.graph.drawing").
		Int("edges", len(edges)).
		Str("title", shared.ReplaceNonAsciiWithSpace(rootChannel.TgTitle)).
		Send()

	drawInput := ds_supplier.DrawGraphInput{
		DrawInput: ds_supplier.DrawInput{
			Title: r.resourceService.Localizef(
				ctx,
				resource.AnaliseChartChannelNeighbors,
				input.Locale,
				rootChannel.TgTitle,
			),
			FigSize:     []int{50, 35},
			ImageFormat: "png",
		},
		Edges:         edges,
		Layout:        "circular_tree",
		WeightedEdges: true,
		Nodes:         nodes,
		RootNode:      shared.ReplaceNonAsciiWithSpace(rootChannel.TgTitle),
	}

	res, err := r.dsSupplier.DrawGraph(ctx, &drawInput)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to draw graph")
	}

	return File{
		Name:      "channel_similarity_graph",
		Extension: "png",
		Content:   res,
	}, nil
}
