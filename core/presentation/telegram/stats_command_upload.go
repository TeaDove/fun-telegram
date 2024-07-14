package telegram

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/guregu/null/v5"
	"github.com/teadove/fun_telegram/core/service/analitics"
	"github.com/teadove/fun_telegram/core/service/resource"
	"github.com/teadove/fun_telegram/core/shared"

	"github.com/celestix/gotgproto/types"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	iterHistoryBatchSize = 100
)

var (
	FlagUploadStatsOffset = optFlag{
		Long:        "offset",
		Short:       "o",
		Description: resource.CommandStatsFlagChannelOffsetDescription,
	}
	FlagUploadStatsDay = optFlag{
		Long:        "day",
		Short:       "d",
		Description: resource.CommandStatsFlagDayDescription,
	}
	FlagUploadStatsRemove = optFlag{
		Long:        "rm",
		Short:       "r",
		Description: resource.CommandStatsFlagRemoveDescription,
	}
	FlagUploadStatsCount = optFlag{
		Long:        "count",
		Short:       "c",
		Description: resource.CommandStatsFlagCountDescription,
	}
	FlagStatsChannelName = optFlag{
		Long:        "channel",
		Short:       "h",
		Description: resource.CommandStatsFlagChannelNameDescription,
	}
	FlagStatsChannelDepth = optFlag{
		Long:        "depth",
		Short:       "p",
		Description: resource.CommandStatsFlagDepthDescription,
	}
	FlagStatsChannelMaxOrder = optFlag{
		Long:        "max-order",
		Short:       "m",
		Description: resource.CommandStatsFlagChannelMaxOrderDescription,
	}
)

func (r *Presentation) uploadMembers(
	ctx context.Context,
	wg *sync.WaitGroup,
	chat types.EffectiveChat,
) {
	defer wg.Done()

	_, err := r.updateMembers(ctx, chat)
	if err != nil {
		zerolog.Ctx(ctx).
			Error().
			Stack().
			Err(errors.WithStack(err)).
			Str("status", "failed.to.update.members").
			Send()

		return
	}
}

func (r *Presentation) uploadMessageToRepository(
	ctx *ext.Context,
	wg *sync.WaitGroup,
	update *ext.Update,
	elemChan <-chan messages.Elem,
) {
	defer wg.Done()

	var elem messages.Elem
	for elem = range elemChan {
		msg, ok := elem.Msg.(*tg.Message)
		if !ok {
			continue
		}

		msgFrom, ok := msg.FromID.(*tg.PeerUser)
		if !ok {
			continue
		}

		analiticsMessage := analitics.Message{
			CreatedAt: time.Unix(int64(msg.Date), 0),
			TgChatID:  update.EffectiveChat().GetID(),
			TgUserId:  msgFrom.UserID,
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

			continue
		}

		zerolog.Ctx(ctx).Trace().Str("status", "message.uploaded").Int("msg_id", msg.ID).Send()
	}
}

func (r *Presentation) uploadStatsDeleteMessages(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	if update.EffectiveChat().GetID() == ctx.Self.ID {
		output, err := r.jobService.DeleteOldMessages(ctx)
		if err != nil {
			return errors.WithStack(err)
		}

		if output.OldCount == 0 {
			err = r.replyIfNotSilent(
				ctx,
				update,
				input,
				"No need to delete messages",
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		err = r.replyIfNotSilent(
			ctx,
			update,
			input,
			fmt.Sprintf(
				"Messages deleted\n"+
					"Old count: %d, New count: %d\n"+
					"Old size: %.2fkb, new size: %.2fkb\n"+
					"Mem freed: %.2fkb",
				output.OldCount, output.NewCount,
				shared.ToKilo(output.OldSize), shared.ToKilo(output.NewSize),
				shared.ToKilo(output.BytesFreed)),
		)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	count, err := r.analiticsService.DeleteMessagesByChatId(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	if !input.Silent {
		_, err = ctx.Reply(update, fmt.Sprintf("%d messages was deleted", count), nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *Presentation) updateUploadStatsMessage(
	ctx *ext.Context,
	count int,
	chatId int64,
	msgId int,
	chatPeer tg.InputPeerClass,
	offset int,
	startedAt time.Time,
	lastDate time.Time,
	maxCount int,
) {
	zerolog.Ctx(ctx).Info().Str("status", "messages.batch.uploaded").Int("count", count).Send()

	elapsed := time.Since(startedAt).Seconds()
	remainingCount := maxCount - count
	speedSeconds := float64(count) / elapsed

	_, err := ctx.EditMessage(chatId, &tg.MessagesEditMessageRequest{
		Peer: chatPeer,
		ID:   msgId,
		Message: fmt.Sprintf(
			`⚙️ Uploading messages

Amount uploaded: %d, Remaining: %d
Seconds elapsed: %.2f, Speed: %.2fmsg/s, ETA: %.1f minutes
Offset: %d
LastDate: %s`,
			count,
			remainingCount,
			elapsed,
			speedSeconds,
			float64(remainingCount)/speedSeconds/60,
			offset,
			lastDate.String(),
		),
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
	}
}

// uploadStatsUpload
func (r *Presentation) uploadStatsUpload( // nolint: cyclop
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) (err error) {
	const (
		maxElapsed = time.Hour
	)

	maxCount := shared.DefaultUploadCount

	if userMaxCountS, ok := input.Ops[FlagUploadStatsCount.Long]; ok {
		userMaxCount, err := strconv.Atoi(userMaxCountS)
		if err != nil {
			_, err := ctx.Reply(
				update,
				fmt.Sprintf("Err: failed to parse count flag: %s", err.Error()),
				nil,
			)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userMaxCount < shared.MaxUploadCount {
			maxCount = userMaxCount
		} else {
			maxCount = shared.MaxUploadCount
		}
	}

	maxQueryAge := shared.DefaultUploadQueryAge

	if userQueryAgeS, ok := input.Ops[FlagUploadStatsDay.Long]; ok {
		userQueryAge, err := strconv.Atoi(userQueryAgeS)
		if err != nil {
			_, err = ctx.Reply(
				update,
				fmt.Sprintf("Err: failed to parse age flag: %s", err.Error()),
				nil,
			)
			if err != nil {
				return errors.WithStack(err)
			}
		}

		if userQueryAge < int(shared.MaxUploadQueryAge.Hours()/24) {
			maxQueryAge = time.Hour * 24 * time.Duration(userQueryAge)
		} else {
			maxQueryAge = shared.MaxUploadQueryAge
		}
	}

	queryTill := time.Now().UTC().Add(-maxQueryAge)

	var (
		barChatId    int64
		barMessageId int
		barPeer      tg.InputPeerClass
	)

	if !input.Silent {
		barMessage, err := ctx.Reply(update, "⚙️ Uploading messages", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		barMessageId = barMessage.ID
		barChatId = update.EffectiveChat().GetID()
		barPeer = update.EffectiveChat().GetInputPeer()
	} else {
		barMessage, err := ctx.SendMessage(ctx.Self.ID, &tg.MessagesSendMessageRequest{Message: "⚙️ Uploading messages"})
		if err != nil {
			return errors.WithStack(err)
		}

		barMessageId = barMessage.ID
		barChatId = ctx.Self.ID
		barPeer = ctx.Self.AsInputPeer()
	}

	offset := 0
	if flaggedOffset, ok := input.Ops[FlagUploadStatsOffset.Long]; ok {
		offset, err = strconv.Atoi(flaggedOffset)
		if err != nil {
			err = r.replyIfNotSilentLocalizedf(
				ctx,
				update,
				input,
				resource.ErrUnprocessableEntity,
				err.Error(),
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}
	}

	zerolog.Ctx(ctx).Info().Str("status", "stats.upload.begin").Int("offset", offset).Send()
	historyQuery := query.Messages(r.telegramApi).GetHistory(update.EffectiveChat().GetInputPeer())
	historyQuery.BatchSize(iterHistoryBatchSize)
	historyQuery.OffsetID(offset)
	historyIter := historyQuery.Iter()
	startedAt := time.Now()
	count := 0

	var wg sync.WaitGroup

	wg.Add(2)

	go r.uploadMembers(ctx, &wg, update.EffectiveChat())

	elemChan := make(chan messages.Elem, iterHistoryBatchSize*3)

	go r.uploadMessageToRepository(ctx, &wg, update, elemChan)

	var lastDate time.Time

	for {
		zerolog.Ctx(ctx).Trace().Str("status", "new.iteration").Int("offset", offset).Send()
		ok := historyIter.Next(ctx)
		if ok {
			zerolog.Ctx(ctx).Trace().Str("status", "elem.got").Send()

			elem := historyIter.Value()
			offset = elem.Msg.GetID()
			msg, ok := elem.Msg.(*tg.Message)
			if !ok {
				zerolog.Ctx(ctx).Trace().Str("status", "not.an.message").Send()
				continue
			}

			lastDate = time.Unix(int64(msg.Date), 0).In(input.ChatSettings.TimeLoc)

			count++
			elemChan <- elem

			if count%iterHistoryBatchSize == 0 {
				time.Sleep(time.Millisecond * 800)

				go r.updateUploadStatsMessage(
					ctx,
					count,
					barChatId,
					barMessageId,
					barPeer,
					offset,
					startedAt,
					lastDate,
					maxCount,
				)
			}

			if !lastDate.After(queryTill) {
				zerolog.Ctx(ctx).Info().Str("status", "last.in.period.message.found").Send()
				break
			}

			if time.Since(startedAt) > maxElapsed {
				zerolog.Ctx(ctx).Info().Str("status", "iterating.too.long").Send()
				break
			}

			if count > maxCount {
				zerolog.Ctx(ctx).Info().Str("status", "iterating.too.much").Send()
				break
			}

			zerolog.Ctx(ctx).Trace().Str("status", "elem.processed").Send()

			continue
		}

		err = historyIter.Err()
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "all.messages.found").Send()

		break
	}

	zerolog.Ctx(ctx).
		Info().
		Str("status", "waiting.for.uploading.to.repository").
		Int("count", count).
		Send()
	close(elemChan)
	wg.Wait()
	zerolog.Ctx(ctx).Info().Str("status", "messages.uploaded").Int("count", count).Send()

	_, err = ctx.EditMessage(barChatId, &tg.MessagesEditMessageRequest{
		Peer: barPeer,
		ID:   barMessageId,
		Message: fmt.Sprintf(
			"Messages uploaded!\n\n"+
				"Amount: %d\n"+
				"Elapsed: %.2fm\n"+
				"LastDate: %s",
			count,
			time.Since(startedAt).Minutes(),
			lastDate.In(input.ChatSettings.TimeLoc).String(),
		),
	})
	if err != nil {
		zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
	}

	return nil
}

func (r *Presentation) uploadStatsCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	if channel, ok := input.Ops[FlagStatsChannelName.Long]; ok {
		return r.uploadChannelStatsMessages(ctx, update, input, channel)
	}

	if _, ok := input.Ops[FlagUploadStatsRemove.Long]; ok {
		return r.uploadStatsDeleteMessages(ctx, update, input)
	}

	return r.uploadStatsUpload(ctx, update, input)
}
