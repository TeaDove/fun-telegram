package telegramapi

import (
	"fmt"
	"strconv"
	"time"

	"github.com/TeaDove/fun-telegram/internal/services/messageservice"

	"github.com/TeaDove/fun-telegram/internal/utils"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/guregu/null/v5"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog"
)

const (
	iterHistoryBatchSize = 100
)

func (r *Presentation) updateUploadStatsMessage(
	ctx *ext.Context,
	count int,
	chatID int64,
	msgID int,
	chatPeer tg.InputPeerClass,
	offset int,
	startedAt time.Time,
	lastDate time.Time,
	maxCount int,
) {
	zerolog.Ctx(ctx).Info().
		Int("count", count).
		Msg("messages.batch.uploaded")

	elapsed := time.Since(startedAt).Seconds()
	remainingCount := maxCount - count
	speedSeconds := float64(count) / elapsed

	_, err := ctx.EditMessage(chatID, &tg.MessagesEditMessageRequest{
		Peer: chatPeer,
		ID:   msgID,
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

func statsGetArgs(c *Context) (getChatStorageInput, error) {
	var input getChatStorageInput

	const (
		maxElapsed = time.Hour
	)

	input.MaxElapsed = maxElapsed
	input.MaxCount = utils.DefaultUploadCount

	if userMaxCountS, ok := c.Ops[FlagUploadStatsCount.Long]; ok {
		userMaxCount, err := strconv.Atoi(userMaxCountS)
		if err != nil {
			return getChatStorageInput{}, errors.Wrap(err, "failed to parse count flag")
		}

		if userMaxCount < utils.MaxUploadCount {
			input.MaxCount = userMaxCount
		} else {
			input.MaxCount = utils.MaxUploadCount
		}
	}

	maxQueryAge := utils.DefaultUploadQueryAge

	if userQueryAgeS, ok := c.Ops[FlagUploadStatsDay.Long]; ok {
		userQueryAge, err := strconv.Atoi(userQueryAgeS)
		if err != nil {
			return getChatStorageInput{}, errors.Wrap(err, "failed to parse age flag")
		}

		if userQueryAge < int(utils.MaxUploadQueryAge.Hours()/24) {
			maxQueryAge = time.Hour * 24 * time.Duration(userQueryAge)
		} else {
			maxQueryAge = utils.MaxUploadQueryAge
		}
	}

	input.QueryTill = time.Now().UTC().Add(-maxQueryAge)

	return input, nil
}

func (r *Presentation) appendMessage(c *Context, storage *messageservice.Storage, elem messages.Elem) {
	msg, ok := elem.Msg.(*tg.Message)
	if !ok {
		return
	}

	msgFrom, ok := msg.FromID.(*tg.PeerUser)
	if !ok {
		return
	}

	analiticsMessage := messageservice.Message{
		CreatedAt: time.Unix(int64(msg.Date), 0),
		TgChatID:  c.update.EffectiveChat().GetID(),
		TgUserID:  msgFrom.UserID,
		Text:      msg.Message,
		TgID:      msg.ID,
	}

	if msg.ReplyTo != nil {
		messageReplyHeader, ok := msg.ReplyTo.(*tg.MessageReplyHeader)
		if ok {
			analiticsMessage.ReplyToTgMsgID = null.IntFrom(int64(messageReplyHeader.ReplyToMsgID))
		}
	}

	r.analiticsService.AppendMessage(storage, &analiticsMessage)
}

type getChatStorageInput struct {
	MaxElapsed time.Duration
	MaxCount   int
	QueryTill  time.Time
}

func startStatsProgress(c *Context) (int64, int, tg.InputPeerClass, error) {
	var (
		barChatID    int64
		barMessageID int
		barPeer      tg.InputPeerClass
	)

	if !c.Silent {
		barMessage, err := c.extCtx.Reply(c.update, ext.ReplyTextString("⚙️ Uploading messages"), nil)
		if err != nil {
			return 0, 0, nil, errors.WithStack(err)
		}

		barMessageID = barMessage.ID
		barChatID = c.update.EffectiveChat().GetID()
		barPeer = c.update.EffectiveChat().GetInputPeer()
	} else {
		barMessage, err := c.extCtx.SendMessage(
			c.extCtx.Self.ID,
			&tg.MessagesSendMessageRequest{Message: "⚙️ Uploading messages"},
		)
		if err != nil {
			return 0, 0, nil, errors.WithStack(err)
		}

		barMessageID = barMessage.ID
		barChatID = c.extCtx.Self.ID
		barPeer = c.extCtx.Self.AsInputPeer()
	}

	return barChatID, barMessageID, barPeer, nil
}

func (r *Presentation) getChatStorage(c *Context, input *getChatStorageInput) (*messageservice.Storage, error) {
	barChatID, barMessageID, barPeer, err := startStatsProgress(c)
	if err != nil {
		return nil, errors.Wrap(err, "start stats progress")
	}

	zerolog.Ctx(c.extCtx).
		Info().
		Msg("stats.upload.begin")

	historyQuery := query.Messages(r.telegramAPI).GetHistory(c.update.EffectiveChat().GetInputPeer())
	historyQuery.BatchSize(iterHistoryBatchSize)
	historyIter := historyQuery.Iter()
	startedAt := time.Now()
	count := 0

	var lastDate time.Time

	storage := r.analiticsService.NewStorage()

	users, err := r.updateMembers(c.extCtx, c.update.EffectiveChat())
	if err != nil {
		return nil, c.replyWithError(errors.WithStack(err))
	}

	storage.Users = users
	storage.UsersNameGetter = storage.Users.GetNameGetter()

	for historyIter.Next(c.extCtx) {
		elem := historyIter.Value()
		offset := elem.Msg.GetID()

		zerolog.Ctx(c.extCtx).Trace().Int("offset", offset).Msg("new.iteration")

		msg, ok := elem.Msg.(*tg.Message)
		if !ok {
			continue
		}

		lastDate = time.Unix(int64(msg.Date), 0).In(utils.TZTime)

		count++

		r.appendMessage(c, storage, elem)

		if count%iterHistoryBatchSize == 0 {
			time.Sleep(time.Millisecond * 800)

			go r.updateUploadStatsMessage(
				c.extCtx,
				count,
				barChatID,
				barMessageID,
				barPeer,
				offset,
				startedAt,
				lastDate,
				input.MaxCount,
			)
		}

		if !lastDate.After(input.QueryTill) || time.Since(startedAt) > input.MaxElapsed || count > input.MaxCount {
			break
		}
	}

	err = historyIter.Err()
	if err != nil {
		return nil, errors.Wrap(err, "iterate chat history")
	}

	zerolog.Ctx(c.extCtx).
		Info().
		Int("count", count).
		Msg("waiting.for.uploading.to.repository")
	zerolog.Ctx(c.extCtx).Info().Str("status", "messages.uploaded").Int("count", count).Send()

	_, err = c.extCtx.EditMessage(barChatID, &tg.MessagesEditMessageRequest{
		Peer: barPeer,
		ID:   barMessageID,
		Message: fmt.Sprintf(
			"Messages uploaded!\n\n"+
				"Amount: %d\n"+
				"Elapsed: %.2fm\n"+
				"LastDate: %s",
			count,
			time.Since(startedAt).Minutes(),
			lastDate.In(utils.TZTime).String(),
		),
	})
	if err != nil {
		zerolog.Ctx(c.extCtx).Error().Stack().Err(err).Str("status", "failed.to.edit.message").Send()
	}

	return storage, nil
}

// statsCommand.
func (r *Presentation) statsCommand(c *Context) error {
	input, err := statsGetArgs(c)
	if err != nil {
		return errors.WithStack(err)
	}

	storage, err := r.getChatStorage(c, &input)
	if err != nil {
		return errors.Wrap(err, "failed to get chat storage")
	}

	return r.compileStats(c, storage)
}
