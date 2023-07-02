package telegram

import (
	"context"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

func (r *Presentation) helpCommandHandler(
	ctx context.Context,
	entities *tg.Entities,
	update message.AnswerableMessageUpdate,
	m *tg.Message,
) error {
	const helpMessage = "Available commands:\n\n" +
		"<code>!help</code> - get this message\n" +
		"<code>!ping</code> - ping all users\n" +
		"<code>!getMe</code> - get id, username of requested user and group"
	_, err := r.telegramSender.Reply(*entities, update).
		StyledText(ctx, html.String(nil, helpMessage))
	return err
}
