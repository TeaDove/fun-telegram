package telegram

import (
	"github.com/anonyindian/gotgproto/ext"
)

func (r *Presentation) helpCommandHandler(ctx *ext.Context, update *ext.Update) error {
	const helpMessage = "Available commands:\n\n" +
		"<code>!help</code> - get this message\n" +
		"<code>!ping</code> - ping all users\n" +
		"<code>!getMe</code> - get id, username of requested user and group"
	_, err := ctx.Reply(update, helpMessage, nil)
	return err
}
