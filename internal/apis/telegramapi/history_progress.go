package telegramapi

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/cockroachdb/errors"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog"
)

type historyProgress func(ctx context.Context, amount int, stopped string)

func renderHistoryProgress(
	header []styling.StyledTextOption,
	amount int,
	stopped string,
) (string, []tg.MessageEntityClass, error) {
	text := slices.Clone(header)
	if amount >= 0 {
		text = append(
			text,
			styling.Plain("\n\nRead "),
			styling.Code(formatHistoryCount(amount)),
			styling.Plain(" msgs"),
		)
	}

	if stopped != "" {
		text = append(text, styling.Plain("\nStopped: "), styling.Code(stopped))
	}

	builder := entity.Builder{}

	err := styling.Perform(&builder, text...)
	if err != nil {
		return "", nil, errors.Wrap(err, "render history progress")
	}

	message, entities := builder.Complete()

	return message, entities, nil
}

func (c *Context) startHistoryProgress(names []string, options historyOptions) (historyProgress, error) {
	unique := make([]string, 0, len(names))
	for _, name := range names {
		if !slices.ContainsFunc(unique, func(item string) bool { return strings.EqualFold(item, name) }) {
			unique = append(unique, name)
		}
	}

	span := formatHistoryDuration(c.StartedAt.Sub(options.cutoff))
	header := []styling.StyledTextOption{
		styling.Plain("Loading history for "), styling.Italic(strings.Join(unique, ", ")),
		styling.Plain(fmt.Sprintf(" with %s amount and %s limit", formatHistoryCount(options.count), span)),
	}

	text, entities, err := renderHistoryProgress(header, -1, "")
	if err != nil {
		return nil, err
	}

	origin := c.update.EffectiveChat().GetInputPeer()

	var replyID int

	if c.Silent || c.inSavedMessages() {
		origin = &tg.InputPeerSelf{}

		sent, err := c.extCtx.SendMessage(c.extCtx.Self.ID,
			&tg.MessagesSendMessageRequest{Peer: origin, Message: text, Entities: entities})
		if err != nil {
			return nil, errors.Wrap(err, "send history progress to Saved Messages")
		}

		replyID = sent.ID
	} else {
		sent, err := c.extCtx.Reply(c.update, ext.ReplyTextStyledTextArray(header), nil)
		if err != nil {
			return nil, errors.Wrap(err, "send history progress")
		}

		replyID = sent.ID
	}

	return c.editHistoryProgress(header, origin, replyID), nil
}

func (c *Context) editHistoryProgress(
	header []styling.StyledTextOption,
	origin tg.InputPeerClass,
	replyID int,
) historyProgress {
	const minInterval = 2 * time.Second

	lastUpdate := time.Now()
	lastMessage := ""

	return func(ctx context.Context, amount int, stopped string) {
		if stopped == "" && time.Since(lastUpdate) < minInterval {
			return
		}

		text, entities, err := renderHistoryProgress(header, amount, stopped)
		if err != nil {
			zerolog.Ctx(c.extCtx).Warn().Err(err).Msg("history.progress.render.failed")
			return
		}

		if text == lastMessage {
			return
		}

		lastUpdate = time.Now()

		_, err = c.extCtx.Raw.MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
			Peer: origin, ID: replyID, Message: text, Entities: entities,
		})
		if err != nil {
			zerolog.Ctx(c.extCtx).Warn().Err(err).Msg("history.progress.edit.failed")
			return
		}

		lastMessage = text
	}
}
