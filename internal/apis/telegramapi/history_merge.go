package telegramapi

import (
	"context"
	"fmt"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/cockroachdb/errors"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
)

type historySource struct {
	chat historyChat
	iter historyIterator
	elem messages.Elem
	msg  *tg.Message
}

func (s *historySource) advance(ctx context.Context) error {
	s.msg = nil

	for ctx.Err() == nil && s.iter.Next(ctx) {
		elem := s.iter.Value()

		msg, ok := elem.Msg.(*tg.Message)
		if ok {
			s.elem = elem
			s.msg = msg

			return nil
		}
	}

	return errors.Wrap(errors.Join(ctx.Err(), s.iter.Err()), "iterate chat history")
}

func newestHistorySource(sources []historySource) *historySource {
	var newest *historySource

	for i := range sources {
		source := &sources[i]
		if source.msg == nil {
			continue
		}

		if newest == nil || source.msg.Date > newest.msg.Date {
			newest = source
		}
	}

	return newest
}

func walkHistory(
	ctx context.Context,
	sources []historySource,
	options historyOptions,
	selfID int64,
	consume func(record tgdtos.Message, author tgdtos.Author) error,
	progress func(context.Context, string),
) (historySummary, error) {
	const progressInterval = 5 * historyBatchSize

	summary := historySummary{stop: "end of history"}

	for i := range sources {
		err := sources[i].advance(ctx)
		if err != nil && ctx.Err() == nil {
			return historySummary{}, err
		}
	}

	for ctx.Err() == nil {
		source := newestHistorySource(sources)
		if source == nil {
			return summary, nil
		}

		record, author := decodeHistoryMessage(source.elem, source.msg, selfID, source.chat.forum)
		record.PeerID = int64(source.chat.id)
		date := time.Unix(record.Date, 0).UTC()

		if date.Before(options.cutoff) {
			summary.stop = "date limit"
			return summary, nil
		}

		err := consume(record, author)
		if err != nil && ctx.Err() == nil {
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

		err = source.advance(ctx)
		if err != nil && ctx.Err() == nil {
			return historySummary{}, err
		}
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		summary.stop = "time limit"
		return summary, nil
	}

	return summary, errors.Wrap(ctx.Err(), "read history")
}
