package analitics

import (
	"archive/zip"
	"bytes"
	"context"
	"github.com/gocarina/gocsv"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/ch_repository"
)

func (r *Service) ChannelInsert(ctx context.Context, channel *ch_repository.Channel) error {
	channel.UploadedAt = time.Now().UTC()

	err := r.chRepository.ChannelInsert(ctx, channel)
	if err != nil {
		return errors.Wrap(err, "failed to insert channel")
	}

	zerolog.Ctx(ctx).Debug().Str("status", "channel.inserted").Interface("title", channel.TgTitle).Send()

	return nil
}

func (r *Service) ChannelSelect(ctx context.Context, id int64) (ch_repository.Channel, error) {
	channel, err := r.chRepository.ChannelSelectById(ctx, id)
	if err != nil {
		return ch_repository.Channel{}, errors.Wrap(err, "failed to select channel")
	}

	return channel, nil
}

var channelDataTtl = time.Hour * 24 * 60

func (r *Service) channelEdgeBatchInsert(ctx context.Context, channels ch_repository.Channels) error {
	edges := make([]ch_repository.ChannelEdge, 0, len(channels)*2)
	for _, channelIn := range channels {
		for idx, channelOut := range channelIn.RecommendationsIds {
			edges = append(edges, ch_repository.ChannelEdge{
				TgIdIn:  channelIn.TgId,
				TgIdOut: channelOut,
				Order:   int64(idx),
			})
		}
	}

	err := r.chRepository.ChannelEdgeBatchInsert(ctx, edges)
	if err != nil {
		return errors.Wrap(err, "failed to batch insert channel edges")
	}

	return nil
}

func (r *Service) ChannelBatchInsert(ctx context.Context, channels []ch_repository.Channel) error {
	channelIds := make([]int64, len(channels))
	for idx, channel := range channels {
		channelIds[idx] = channel.TgId
	}

	oldChannels, err := r.chRepository.ChannelSelectByIds(ctx, channelIds)
	if err != nil {
		return errors.Wrap(err, "failed to select channels")
	}

	oldChannelsMap := oldChannels.ToMap()

	leafsCount := 0
	alreadyProcessed := 0
	newChannels := make([]ch_repository.Channel, 0, len(channels))
	for _, channel := range channels {
		oldChannel, ok := oldChannelsMap[channel.TgId]
		if ok && time.Since(oldChannel.UploadedAt) < channelDataTtl && !oldChannel.IsLeaf && channel.IsLeaf {
			alreadyProcessed++
			continue
		}

		channel.UploadedAt = time.Now().UTC()
		if channel.RecommendationsIds == nil {
			leafsCount++
		}
		newChannels = append(newChannels, channel)
	}

	err = r.chRepository.ChannelBatchInsert(ctx, newChannels)
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

func dumpSliceToCsvZip(name string, slice any) (File, error) {
	file := File{
		Name:      name,
		Extension: "csv.zip",
	}

	sliceBytes, err := gocsv.MarshalBytes(slice)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to dump to csv")
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	zipFile, err := zipWriter.Create(name + ".csv")
	if err != nil {
		return File{}, errors.Wrap(err, "failed to create zip")
	}

	_, err = zipFile.Write(sliceBytes)
	if err != nil {
		return File{}, errors.Wrap(err, "failed to write bytes to zip")
	}

	err = zipWriter.Close()
	if err != nil {
		return File{}, errors.Wrap(err, "failed to close zip writer")
	}

	file.Content = buf.Bytes()

	return file, nil
}

func (r *Service) DumpChannels(ctx context.Context) ([]File, error) {
	files := make([]File, 0, 3)

	channels, err := r.chRepository.ChannelSelect(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select channel")
	}

	file, err := dumpSliceToCsvZip("channels", channels)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dump to csv channels")
	}

	files = append(files, file)

	channelEdges, err := r.chRepository.ChannelEdgesSelect(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select channel edges")
	}

	file, err = dumpSliceToCsvZip("channel_edges", channelEdges)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dump to csv channels")
	}

	files = append(files, file)

	messages, err := r.chRepository.MessagesGetChannel(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select channels messages")
	}

	file, err = dumpSliceToCsvZip("messages", messages)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dump to csv channels")
	}

	files = append(files, file)

	return files, nil
}
