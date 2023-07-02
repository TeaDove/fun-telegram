package telegram

import (
	"context"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/teadove/goteleout/internal/utils"
)

func (r *Presentation) statsCommandHandler(
	ctx context.Context,
	entities *tg.Entities,
	update message.AnswerableMessageUpdate,
	m *tg.Message,
) error {
	var peer tg.InputPeerClass
	for _, value := range entities.Channels {
		peer = value.AsInputPeer()
		break
	}
	for _, value := range entities.Chats {
		peer = value.AsInputPeer()
		break
	}

	err := query.Messages(r.telegramApi).
		GetHistory(peer).
		BatchSize(100).
		ForEach(ctx, func(ctx context.Context, elem messages.Elem) error {
			elemMessage, ok := elem.Msg.(*tg.Message)
			if !ok {
				return nil
			}
			println(elemMessage.Message)
			for key, v := range elem.Entities.Users() {
				println(key)
				utils.LogInterface(v.Username)
			}
			return nil
		})
	if err != nil {
		return err
	}
	_, err = r.telegramSender.Reply(*entities, update).
		StyledText(ctx, html.String(nil, "aaa"))
	return err
}
