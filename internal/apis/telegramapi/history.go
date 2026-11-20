package telegramapi

import (
	"context"
	"strconv"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/TeaDove/fun-telegram/internal/utils"
	"github.com/cockroachdb/errors"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
)

const (
	historyBatchSize    = 100
	historyCountFlag    = "count"
	historyDayFlag      = "day"
	historyDay          = 24 * time.Hour
	historyReadDeadline = time.Hour
)

func historyFlags() []optFlag {
	return []optFlag{
		{
			Long:        historyCountFlag,
			Short:       "c",
			HasValue:    true,
			Description: "maximum total number of messages across all selected chats",
		},
		{Long: historyDayFlag, Short: "d", HasValue: true, Description: "maximum message age in days"},
		{
			Long: historyNameFlag, Short: "n", HasValue: true,
			Description: "select all chats with this name, repeat for multiple names",
		},
		{Long: FlagSilent.Long, Short: FlagSilent.Short, Description: "send progress and results to Saved Messages"},
	}
}

type historyOptions struct {
	count  int
	cutoff time.Time
}

func readHistoryLimit(options map[string][]string, key string, defaultValue, maxValue int) (int, error) {
	raw, ok := options[key]
	if !ok || len(raw) == 0 {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw[len(raw)-1])
	if err != nil {
		return 0, errors.Wrapf(err, "parse --%s", key)
	}

	if value < 1 {
		return 0, errors.Newf("--%s must be a positive integer", key)
	}

	return min(value, maxValue), nil
}

func parseHistoryOptions(options map[string][]string, now time.Time) (historyOptions, error) {
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

	record := tgdtos.Message{
		ID: msg.ID, PeerID: int64(telegramPeerID(msg.PeerID)),
		Date: int64(msg.Date), Text: msg.Message, AuthorID: author.ID,
	}
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

func (r *Presentation) readHistory(
	c *Context,
	request historyRequest,
	consume func(record tgdtos.Message, author tgdtos.Author) error,
) (historySummary, error) {
	ctx, cancel := context.WithTimeout(c.extCtx, historyReadDeadline)
	defer cancel()

	sources := make([]historySource, 0, len(request.chats))
	for _, chat := range request.chats {
		history := query.Messages(r.telegramAPI).GetHistory(chat.peer).BatchSize(historyBatchSize)
		if !c.inSavedMessages() && chat.id == telegramPeerID(c.update.EffectiveMessage.PeerID) {
			history.OffsetID(c.update.EffectiveMessage.ID)
		}

		sources = append(sources, historySource{chat: chat, iter: history.Iter()})
	}

	summary, err := walkHistory(ctx, sources, request.options, c.extCtx.Self.ID, consume, request.progress)
	if err != nil {
		request.progress(c.extCtx, summary.count, "error")
		return summary, err
	}

	request.progress(c.extCtx, summary.count, summary.stop)

	return summary, nil
}

type historyIterator interface {
	Next(context.Context) bool
	Value() messages.Elem
	Err() error
}
