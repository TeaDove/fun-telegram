package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/gotd/td/telegram/message/html"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

func (r *Presentation) getMeCommandHandler(
	ctx context.Context,
	entities *tg.Entities,
	update message.AnswerableMessageUpdate,
	m *tg.Message,
) error {
	var textBuilder strings.Builder
	const requestedUserTmp = "Requested user: \n" +
		"id: <code>%d</code>\n" +
		"username: @%s\n\n"
	const currentChatTmp = "Current chat: \n" +
		"id: <code>%d</code>"
	for _, user := range entities.Users {
		textBuilder.WriteString(fmt.Sprintf(requestedUserTmp, user.ID, user.Username))
		break
	}
	for _, chat := range entities.Chats {
		textBuilder.WriteString(fmt.Sprintf(currentChatTmp, chat.ID))
		break
	}
	for _, chat := range entities.Channels {
		textBuilder.WriteString(fmt.Sprintf(currentChatTmp, chat.ID))
		break
	}
	_, err := r.telegramSender.Reply(*entities, update).
		StyledText(ctx, html.String(nil, textBuilder.String()))
	return err
}
