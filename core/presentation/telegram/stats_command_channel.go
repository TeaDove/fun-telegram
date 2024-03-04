package telegram

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/ch_repository"
)

func (r *Presentation) getChannelRecommendations(ctx context.Context, channel *tg.Channel) ([]tg.Channel, error) {
	inputPeerChannel := tg.InputChannel{
		ChannelID:  channel.ID,
		AccessHash: channel.AccessHash,
	}

	t0 := time.Now()

	recommendations, err := r.telegramApi.ChannelsGetChannelRecommendations(ctx, &inputPeerChannel)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get channel recommendations")
	}

	chats := recommendations.GetChats()

	channels := make([]tg.Channel, 0, len(chats))
	for _, recommendedChat := range chats {
		recommendedChannel, ok := recommendedChat.(*tg.Channel) // nolint: sloppyTypeAssert
		if !ok {
			zerolog.Ctx(ctx).Error().
				Str("status", "not.channel.in.recommended.chat").
				Interface("recommended.chat", recommendedChat).
				Interface("original.channel", channel).
				Send()
			continue
		}
		channels = append(channels, *recommendedChannel)
	}

	zerolog.Ctx(ctx).Debug().
		Str("status", "recommendations.got").
		Str("title", channel.Title).
		Int("count", len(channels)).
		Dur("elapsed", time.Since(t0)).
		Send()
	time.Sleep(500 * time.Millisecond)

	return channels, nil
}

func (r *Presentation) dumpChannelRecommendations(
	ctx *ext.Context,
	channels *map[int64]ch_repository.Channel,
	input *input,
	barMessageId int,
	update *ext.Update,
	chat *tg.Channel,
	depth int,
	stopRecursion bool,
	maxDepth int,
	maxRecommendation int,
) error {
	depth++

	foundChannel, ok := (*channels)[chat.ID]
	if ok && len(foundChannel.RecommendationsIds) != 0 {
		zerolog.Ctx(ctx).Trace().Str("status", "channel.already.processed").Str("title", chat.Title).Send()
		return nil
	}

	repositoryChannel := ch_repository.Channel{
		TgId:             chat.ID,
		TgTitle:          chat.Title,
		TgUsername:       chat.Username,
		ParticipantCount: int64(chat.ParticipantsCount),
	}

	if depth > maxDepth || stopRecursion {
		(*channels)[chat.ID] = repositoryChannel

		return nil
	}

	if len(*channels)%10 == 0 {
		go r.updateUploadChannelStatsMessage(ctx, update, barMessageId, input, len(*channels))
	}

	recommendedChannels, err := r.getChannelRecommendations(ctx, chat)
	if err != nil {
		return errors.Wrap(err, "failed to get channel recommendations")
	}

	repositoryChannel.RecommendationsIds = make([]int64, 0, len(recommendedChannels))
	for _, recommendedChannel := range recommendedChannels {
		repositoryChannel.RecommendationsIds = append(repositoryChannel.RecommendationsIds, recommendedChannel.ID)
	}

	(*channels)[chat.ID] = repositoryChannel

	stopRecursion = false
	for idx, recommendedChannel := range recommendedChannels {
		if !stopRecursion && idx > maxRecommendation {
			stopRecursion = true
		}

		err = r.dumpChannelRecommendations(ctx, channels, input, barMessageId, update, &recommendedChannel, depth, stopRecursion, maxDepth, maxRecommendation)
		if err != nil {
			return errors.Wrap(err, "failed to dump nested channels")
		}
	}

	return nil
}

const (
	defaultRecommendation    = 10
	allowedMaxRecommendation = 50

	defaultMaxDepth = 3
	allowedMaxDepth = 10
)

func (r *Presentation) populateChannel(ctx context.Context, channel *tg.Channel) error {
	if channel.ParticipantsCount != 0 {
		return nil
	}

	fullChannelClass, err := r.telegramApi.ChannelsGetFullChannel(ctx, channel.AsInput())
	if err != nil {
		return errors.Wrap(err, "failed to get full channel")
	}

	fullChannel, ok := fullChannelClass.FullChat.(*tg.ChannelFull) // nolint: sloppyTypeAssert
	if !ok {
		return errors.New("not a channel")
	}

	channel.ParticipantsCount = fullChannel.ParticipantsCount
	return nil
}

func (r *Presentation) updateUploadChannelStatsMessage(ctx *ext.Context, update *ext.Update, msgId int, input *input, count int) {
	_, err := ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		Peer: update.EffectiveChat().GetInputPeer(),
		ID:   msgId,
		Message: fmt.Sprintf(
			"Channels uploading\n\n"+
				"Amount: %d\n"+
				"Elapsed: %.2fm\n",
			count,
			time.Since(input.StartedAt).Minutes(),
		),
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
	}
}

func (r *Presentation) uploadChannelStatsMessages(ctx *ext.Context, update *ext.Update, input *input, channelName string) error {
	var maxDepth = defaultMaxDepth
	if userFlagS, ok := input.Ops[FlagUploadStatsDepth.Long]; ok {
		userV, err := strconv.Atoi(userFlagS)
		if err != nil {
			_, err = ctx.Reply(update, fmt.Sprintf("Err: failed to parse max depth flag: %s", err.Error()), nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userV < allowedMaxDepth {
			maxDepth = userV
		} else {
			maxDepth = allowedMaxDepth
		}
	}

	var maxRecommendation = defaultRecommendation
	if userFlagS, ok := input.Ops[FlagUploadStatsMaxRecommendations.Long]; ok {
		userV, err := strconv.Atoi(userFlagS)
		if err != nil {
			_, err = ctx.Reply(update, fmt.Sprintf("Err: failed to parse max recommendation flag: %s", err.Error()), nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userV < allowedMaxRecommendation {
			maxDepth = userV
		} else {
			maxDepth = allowedMaxRecommendation
		}
	}

	channel, err := r.telegramApi.ContactsResolveUsername(ctx, channelName)
	if err != nil {
		return errors.Wrap(err, "failed to resolve channel")
	}

	if len(channel.Chats) < 1 {
		return errors.New("chats not found")
	}

	realChannel, ok := channel.Chats[0].(*tg.Channel) // nolint: sloppyTypeAssert
	if !ok {
		return errors.New("not an channel")
	}

	err = r.populateChannel(ctx, realChannel)
	if err != nil {
		return errors.Wrap(err, "failed to populate channel")
	}

	barMessage, err := ctx.Reply(update, "⚙️ Uploading channels data", nil)
	if err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	channels := make(map[int64]ch_repository.Channel, 1000)

	err = r.dumpChannelRecommendations(
		ctx,
		&channels,
		input,
		barMessage.ID,
		update,
		realChannel,
		0,
		false,
		maxDepth,
		maxRecommendation,
	)
	if err != nil {
		return errors.Wrap(err, "failed to dump channel recommendations")
	}

	channelsSlice := make([]ch_repository.Channel, 0, len(channels))
	for _, dumpedChannel := range channels {
		channelsSlice = append(channelsSlice, dumpedChannel)
	}

	err = r.analiticsService.ChannelBatchInsert(ctx, channelsSlice)
	if err != nil {
		return errors.Wrap(err, "failed to batch insert channels")
	}

	_, err = ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		Peer: update.EffectiveChat().GetInputPeer(),
		ID:   barMessage.ID,
		Message: fmt.Sprintf(
			"Channels uploaded!\n\n"+
				"Amount: %d\n"+
				"Elapsed: %.2fm\n",
			len(channelsSlice),
			time.Since(input.StartedAt).Minutes(),
		),
	})

	return nil
}
