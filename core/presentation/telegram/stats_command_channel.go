package telegram

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/guregu/null/v5"
	"github.com/teadove/fun_telegram/core/service/analitics"
	"github.com/teadove/fun_telegram/core/shared"
	"strconv"
	"sync"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/ch_repository"
)

const (
	channelsDumpBatch = 100
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

func (r *Presentation) loadChannelMessage(ctx context.Context, wg *sync.WaitGroup, chat *tg.Channel, elem messages.Elem) {
	defer wg.Done()

	msg, ok := elem.Msg.(*tg.Message)
	if !ok {
		return
	}

	if msg.Message == "" {
		return
	}

	analiticsMessage := analitics.Message{
		CreatedAt: time.Unix(int64(msg.Date), 0),
		TgChatID:  chat.ID,
		TgUserId:  chat.ID,
		Text:      msg.Message,
		TgId:      int64(msg.ID),
	}

	if msg.ReplyTo != nil {
		messageReplyHeader, ok := msg.ReplyTo.(*tg.MessageReplyHeader)
		if ok {
			analiticsMessage.ReplyToMsgID = null.IntFrom(int64(messageReplyHeader.ReplyToMsgID))
		}
	}

	err := r.analiticsService.MessageInsert(ctx, &analiticsMessage)
	if err != nil {
		zerolog.Ctx(ctx).
			Error().
			Stack().
			Err(errors.WithStack(err)).
			Str("status", "failed.to.upload.message.to.repository").
			Send()
	}
}

func (r *Presentation) loadChannelMessages(ctx context.Context, chat *tg.Channel) error {
	historyQuery := query.Messages(r.telegramApi).GetHistory(chat.AsInputPeer())
	historyQuery.BatchSize(iterHistoryBatchSize)
	historyQuery.OffsetID(0)
	historyIter := historyQuery.Iter()

	msgCount := 0
	const maxMsgCount = 80

	elemCount := 0
	const maxElemCount = 300

	var wg sync.WaitGroup

	for {
		ok := historyIter.Next(ctx)
		if ok {
			elem := historyIter.Value()
			elemCount++

			msg, ok := elem.Msg.(*tg.Message)
			if !ok {
				continue
			}

			if msg.Message == "" {
				continue
			}

			msgCount++
			wg.Add(1)
			go r.loadChannelMessage(ctx, &wg, chat, elem)

			if msgCount >= maxMsgCount {
				break
			}

			if elemCount >= maxElemCount {
				break
			}

			continue
		}

		err := historyIter.Err()
		if err != nil {
			return errors.Wrap(err, "failed to iterate")
		}

		break
	}

	wg.Wait()
	zerolog.Ctx(ctx).
		Debug().
		Str("status", "channel.messages.uploaded").
		Int("msg.count", msgCount).
		Int("elem.count", elemCount).
		Str("title", chat.Title).
		Send()
	return nil
}

func (r *Presentation) uploadChannelsToRepository(ctx context.Context, channels <-chan Channel) error {
	channelsSlice := make([]ch_repository.Channel, 0, channelsDumpBatch)

	for channel := range channels {
		var tgAbout *string
		if channel.TgAbout.Valid {
			tgAbout = &channel.TgAbout.String
		}
		channelsSlice = append(channelsSlice, ch_repository.Channel{
			TgId:               channel.TgId,
			TgTitle:            channel.TgTitle,
			TgUsername:         channel.TgUsername,
			ParticipantCount:   channel.ParticipantCount,
			RecommendationsIds: channel.RecommendationsIds,
			IsLeaf:             channel.IsLeaf,
			TgAbout:            tgAbout,
		})

		if len(channelsSlice) >= channelsDumpBatch {
			err := r.analiticsService.ChannelBatchInsert(ctx, channelsSlice)
			if err != nil {
				return errors.Wrap(err, "failed to to batch insert")
			}

			channelsSlice = make([]ch_repository.Channel, 0, channelsDumpBatch)
		}
	}

	if len(channelsSlice) != 0 {
		err := r.analiticsService.ChannelBatchInsert(ctx, channelsSlice)
		if err != nil {
			return errors.Wrap(err, "failed to to batch insert")
		}
	}

	return nil
}

type dumpChannelRecommendationsInput struct {
	tgInput      *input
	channels     *map[int64]Channel
	channelsChan chan<- Channel
	update       *ext.Update
	chat         *tg.Channel

	depth         int
	stopRecursion bool
	path          []string

	maxDepth          int
	maxRecommendation int
	barMessageId      int
}

func (r *Presentation) dumpChannelRecommendations(
	ctx *ext.Context,
	input dumpChannelRecommendationsInput,
) error {
	input.depth++

	foundChannel, ok := (*input.channels)[input.chat.ID]
	if ok && !foundChannel.IsLeaf {
		if input.depth < foundChannel.Depth {
			// TODO optimise
			zerolog.Ctx(ctx).Debug().Str("status", "channel.already.processed.but.with.less.depth").Str("title", input.chat.Title).Send()
		} else {
			zerolog.Ctx(ctx).Trace().Str("status", "channel.already.processed").Str("title", input.chat.Title).Send()
			return nil
		}
	}

	repositoryChannel := Channel{
		TgId:             input.chat.ID,
		TgTitle:          shared.ReplaceNonAsciiWithSpace(input.chat.Title),
		TgUsername:       input.chat.Username,
		ParticipantCount: int64(input.chat.ParticipantsCount),
		Depth:            input.depth,
		IsLeaf:           true,
	}

	if input.depth > input.maxDepth || input.stopRecursion {
		(*input.channels)[input.chat.ID] = repositoryChannel
		input.channelsChan <- repositoryChannel

		return nil
	}

	repositoryChannel.IsLeaf = false

	fullChannel, err := r.getFullChannel(ctx, input.chat)
	if err != nil {
		return errors.Wrap(err, "failed to populate channel")
	}

	repositoryChannel.ParticipantCount = int64(fullChannel.ParticipantsCount)
	repositoryChannel.TgAbout = null.StringFrom(fullChannel.About)

	recommendedChannels, err := r.getChannelRecommendations(ctx, input.chat)
	if err != nil {
		return errors.Wrap(err, "failed to get channel recommendations")
	}

	repositoryChannel.RecommendationsIds = make([]int64, 0, len(recommendedChannels))
	for _, recommendedChannel := range recommendedChannels {
		repositoryChannel.RecommendationsIds = append(repositoryChannel.RecommendationsIds, recommendedChannel.ID)
	}

	err = r.loadChannelMessages(ctx, input.chat)
	if err != nil {
		return errors.Wrap(err, "failed to load channel messages")
	}

	go r.updateUploadChannelStatsMessage(ctx, input.update, input.barMessageId, input.tgInput, len(*input.channels), &repositoryChannel, input.path)

	(*input.channels)[input.chat.ID] = repositoryChannel
	input.channelsChan <- repositoryChannel

	input.stopRecursion = false
	for idx, recommendedChannel := range recommendedChannels {
		if !input.stopRecursion && idx > input.maxRecommendation {
			input.stopRecursion = true
		}

		newPath := make([]string, len(input.path))
		_ = copy(newPath, input.path)
		newPath = append(newPath, fmt.Sprintf("(%d) %s (@%s)", idx, recommendedChannel.Title, recommendedChannel.Username))

		err = r.dumpChannelRecommendations(
			ctx,
			dumpChannelRecommendationsInput{
				tgInput:           input.tgInput,
				channels:          input.channels,
				barMessageId:      input.barMessageId,
				update:            input.update,
				chat:              &recommendedChannel,
				depth:             input.depth,
				stopRecursion:     input.stopRecursion,
				maxDepth:          input.maxDepth,
				maxRecommendation: input.maxRecommendation,
				path:              newPath,
				channelsChan:      input.channelsChan,
			},
		)
		if err != nil {
			return errors.Wrap(err, "failed to dump nested channels")
		}
	}

	return nil
}

const (
	defaultOrder    = 10
	allowedMaxOrder = 50

	defaultMaxDepth = 3
	allowedMaxDepth = 10
)

func (r *Presentation) getFullChannel(ctx context.Context, channel *tg.Channel) (*tg.ChannelFull, error) {
	fullChannelClass, err := r.telegramApi.ChannelsGetFullChannel(ctx, channel.AsInput())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get full channel")
	}

	fullChannel, ok := fullChannelClass.FullChat.(*tg.ChannelFull) // nolint: sloppyTypeAssert
	if !ok {
		return nil, errors.New("not a channel")
	}

	return fullChannel, nil
}

func (r *Presentation) updateUploadChannelStatsMessage(
	ctx *ext.Context,
	update *ext.Update,
	msgId int,
	input *input,
	count int,
	channel *Channel,
	path []string,
) {
	var pathText string
	for _, pathItem := range path {
		pathText += "\n ➔ " + pathItem
	}

	elapsed := time.Since(input.StartedAt).Minutes()
	_, err := ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		Peer: update.EffectiveChat().GetInputPeer(),
		ID:   msgId,
		Message: fmt.Sprintf(
			"Channels uploading\n\n"+
				"Recommendations found for current channel: %d\n"+
				"Total channels found: %d\n"+
				"Elapsed: %.2fm, Speed: %.2f(channels/m)\n"+
				"Path: %s\n",
			len(channel.RecommendationsIds),
			count,
			elapsed,
			float64(count)/elapsed,
			pathText,
		),
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
	}
}

type Channel struct {
	TgId       int64
	TgTitle    string
	TgAbout    null.String
	IsLeaf     bool
	TgUsername string

	ParticipantCount   int64
	RecommendationsIds []int64

	Depth int
}

func (r *Presentation) uploadChannelStatsMessages(ctx *ext.Context, update *ext.Update, input *input, channelName string) error {
	var maxDepth = defaultMaxDepth
	if userFlagS, ok := input.Ops[FlagStatsChannelDepth.Long]; ok {
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

	var maxRecommendation = defaultOrder
	if userFlagS, ok := input.Ops[FlagStatsChannelMaxOrder.Long]; ok {
		userV, err := strconv.Atoi(userFlagS)
		if err != nil {
			_, err = ctx.Reply(update, fmt.Sprintf("Err: failed to parse max recommendation flag: %s", err.Error()), nil)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userV < allowedMaxOrder {
			maxRecommendation = userV
		} else {
			maxRecommendation = allowedMaxOrder
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

	barMessage, err := ctx.Reply(update, "⚙️ Uploading channels data", nil)
	if err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	channels := make(map[int64]Channel, 1000)
	channelsChan := make(chan Channel, channelsDumpBatch*2)
	var wg sync.WaitGroup

	wg.Add(1)

	var uploadChannelsToRepositoryErr error
	go func() {
		defer wg.Done()
		uploadChannelsToRepositoryErr = r.uploadChannelsToRepository(ctx, channelsChan)
	}()

	err = r.dumpChannelRecommendations(
		ctx,
		dumpChannelRecommendationsInput{
			tgInput:           input,
			channels:          &channels,
			barMessageId:      barMessage.ID,
			update:            update,
			chat:              realChannel,
			depth:             0,
			stopRecursion:     false,
			maxDepth:          maxDepth,
			maxRecommendation: maxRecommendation,
			path:              []string{realChannel.Title},
			channelsChan:      channelsChan,
		},
	)
	if err != nil {
		return errors.Wrap(err, "failed to dump channel recommendations")
	}

	close(channelsChan)
	wg.Wait()

	if uploadChannelsToRepositoryErr != nil {
		return errors.Wrap(uploadChannelsToRepositoryErr, "failed to upload channels to repository")
	}

	_, err = ctx.EditMessage(update.EffectiveChat().GetID(), &tg.MessagesEditMessageRequest{
		Peer: update.EffectiveChat().GetInputPeer(),
		ID:   barMessage.ID,
		Message: fmt.Sprintf(
			"Channels uploaded!\n\n"+
				"Amount: %d\n"+
				"Elapsed: %.2fm\n",
			len(channels),
			time.Since(input.StartedAt).Minutes(),
		),
	})

	return nil
}
