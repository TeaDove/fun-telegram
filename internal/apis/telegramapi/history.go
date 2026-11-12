package telegramapi

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/TeaDove/fun-telegram/internal/utils"
	"github.com/celestix/gotgproto/ext"
	"github.com/cockroachdb/errors"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog"
)

const (
	historyBatchSize    = 100
	historyCountFlag    = "count"
	historyDayFlag      = "day"
	historyDay          = 24 * time.Hour
	historyRangeFormat  = "2006-01-02 15:04:05 UTC"
	historyReadDeadline = time.Hour
)

func historyFlags() []optFlag {
	return []optFlag{
		{Long: historyCountFlag, Short: "c", Description: "maximum number of messages"},
		{Long: historyDayFlag, Short: "d", Description: "maximum message age in days"},
		{Long: historyPeerIDFlag, Description: "select a chat by peer_id from Saved Messages"},
		{Long: FlagSilent.Long, Short: FlagSilent.Short, Description: "send progress and results to Saved Messages"},
	}
}

type historyOptions struct {
	count  int
	cutoff time.Time
}

func readHistoryLimit(options map[string]string, key string, defaultValue, maxValue int) (int, error) {
	raw, ok := options[key]
	if !ok {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, errors.Wrapf(err, "parse --%s", key)
	}

	if value < 1 {
		return 0, errors.Newf("--%s must be a positive integer", key)
	}

	return min(value, maxValue), nil
}

func parseHistoryOptions(options map[string]string, now time.Time) (historyOptions, error) {
	count, err := readHistoryLimit(options, historyCountFlag, utils.DefaultUploadCount, utils.MaxUploadCount)
	if err != nil {
		return historyOptions{}, err
	}

	days, err := readHistoryLimit(options, historyDayFlag,
		int(utils.DefaultUploadQueryAge/historyDay), int(utils.MaxUploadQueryAge/historyDay))
	if err != nil {
		return historyOptions{}, err
	}

	return historyOptions{count: count, cutoff: now.UTC().Add(-time.Duration(days) * historyDay)}, nil
}

func historyAuthor(msg *tg.Message, selfID int64) tg.PeerClass {
	if msg.FromID != nil {
		return msg.FromID
	}

	switch p := msg.PeerID.(type) {
	case *tg.PeerChannel:
		return p
	case *tg.PeerUser:
		if msg.Out {
			return &tg.PeerUser{UserID: selfID}
		}

		return p
	default:
		return nil
	}
}

func decodeHistoryMessage(
	elem messages.Elem,
	msg *tg.Message,
	selfID int64,
	forum bool,
) (tgdtos.Message, tgdtos.Author) {
	p := historyAuthor(msg, selfID)
	author := tgdtos.Author{
		ID:   int64(telegramPeerID(p)),
		Name: historyPeerName(p, elem.Entities),
	}

	if userPeer, ok := p.(*tg.PeerUser); ok {
		if user, found := elem.Entities.User(userPeer.UserID); found {
			author.Username = user.Username
		}
	}

	record := tgdtos.Message{ID: msg.ID, Date: int64(msg.Date), Text: msg.Message, AuthorID: author.ID}
	if forum {
		record.TopicID = 1
	}

	if reply, ok := msg.ReplyTo.(*tg.MessageReplyHeader); ok {
		if reply.ReplyToPeerID == nil || telegramPeerID(reply.ReplyToPeerID) == telegramPeerID(msg.PeerID) {
			record.ReplyTo = reply.ReplyToMsgID
		}

		if reply.ForumTopic {
			record.TopicID = reply.ReplyToTopID
			if record.TopicID == 0 {
				record.TopicID = record.ReplyTo
			}
		}
	}

	return record, author
}

type historySummary struct {
	newest time.Time
	oldest time.Time
	count  int
	stop   string
}

func (s *historySummary) append(date time.Time) {
	if s.count == 0 {
		s.newest = date
	}

	s.oldest = date
	s.count++
}

func (s *historySummary) caption() string {
	if s.count == 0 {
		return fmt.Sprintf("Messages read: 0\nStopped: %s", s.stop)
	}

	return fmt.Sprintf("Messages read: %d\nRange: %s to %s\nStopped: %s",
		s.count, s.oldest.Format(historyRangeFormat), s.newest.Format(historyRangeFormat), s.stop)
}

func (c *Context) startHistoryProgress(chat historyChat) (func(context.Context, string), error) {
	text := chat.label() + "\nReading messages…"
	origin := c.update.EffectiveChat().GetInputPeer()
	replyID := 0

	if c.Silent || c.inSavedMessages() {
		origin = &tg.InputPeerSelf{}

		sent, err := c.extCtx.SendMessage(c.extCtx.Self.ID, &tg.MessagesSendMessageRequest{Peer: origin, Message: text})
		if err != nil {
			return nil, errors.Wrap(err, "send history progress to Saved Messages")
		}

		replyID = sent.ID
	} else {
		sent, err := c.extCtx.Reply(c.update, ext.ReplyTextString(text), nil)
		if err != nil {
			return nil, errors.Wrap(err, "send history progress")
		}

		replyID = sent.ID
	}

	return func(ctx context.Context, status string) {
		_, err := c.extCtx.Raw.MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
			Peer: origin, ID: replyID, Message: chat.label() + "\n" + status,
		})
		if err != nil {
			zerolog.Ctx(c.extCtx).Warn().Err(err).Msg("history.progress.edit.failed")
		}
	}, nil
}

func (r *Presentation) readHistory(
	c *Context,
	chat historyChat,
	consume func(record tgdtos.Message, author tgdtos.Author) error,
) (historySummary, error) {
	options, err := parseHistoryOptions(c.Ops, c.StartedAt)
	if err != nil {
		return historySummary{}, err
	}

	progress, err := c.startHistoryProgress(chat)
	if err != nil {
		return historySummary{}, err
	}

	ctx, cancel := context.WithTimeout(c.extCtx, historyReadDeadline)
	defer cancel()

	history := query.Messages(r.telegramAPI).GetHistory(chat.peer).BatchSize(historyBatchSize)
	if !c.inSavedMessages() {
		history.OffsetID(c.update.EffectiveMessage.ID)
	}

	iter := history.Iter()

	summary, err := walkHistory(ctx, iter, options, c.extCtx.Self.ID, chat.forum, consume, progress)
	if err != nil {
		progress(c.extCtx, "History could not be read.")
		return historySummary{}, err
	}

	progress(c.extCtx, summary.caption())

	return summary, nil
}

type historyIterator interface {
	Next(context.Context) bool
	Value() messages.Elem
	Err() error
}

func walkHistory(
	ctx context.Context,
	iter historyIterator,
	options historyOptions,
	selfID int64,
	forum bool,
	consume func(record tgdtos.Message, author tgdtos.Author) error,
	progress func(context.Context, string),
) (historySummary, error) {
	const progressInterval = 5 * historyBatchSize

	summary := historySummary{stop: "end of history"}

	for ctx.Err() == nil && iter.Next(ctx) {
		elem := iter.Value()

		msg, ok := elem.Msg.(*tg.Message)
		if !ok {
			continue
		}

		record, author := decodeHistoryMessage(elem, msg, selfID, forum)
		date := time.Unix(record.Date, 0).UTC()

		if date.Before(options.cutoff) {
			summary.stop = "date limit"
			return summary, nil
		}

		err := consume(record, author)
		if err != nil {
			return historySummary{}, errors.Wrap(err, "process history message")
		}

		summary.append(date)

		if summary.count == options.count {
			summary.stop = "message limit"
			return summary, nil
		}

		if summary.count%progressInterval == 0 {
			progress(ctx, fmt.Sprintf("Reading messages… %d read", summary.count))
		}
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		summary.stop = "time limit"
		return summary, nil
	}

	err := iter.Err()
	if err != nil {
		return historySummary{}, errors.Wrap(err, "iterate chat history")
	}

	return summary, errors.Wrap(ctx.Err(), "read history")
}
