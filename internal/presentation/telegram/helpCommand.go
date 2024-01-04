package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
)

var helpMessage = []styling.StyledTextOption{
	styling.Plain("Available commands:\n\n"),
	styling.Code("!help"), styling.Plain(" - get this message\n\n"),
	styling.Code("!ping"), styling.Plain(" - ping all users\n\n"),
	styling.Code("!get_me"), styling.Plain(" - get id, username of requested user and group\n\n"),
	styling.Code(
		"!spam_reaction [stop] [disable]",
	), styling.Plain(" - if replied to message with reaction, will spam this reaction to replied user.\n" +
		"[stop] - stop spamming.\n" +
		"[disable] - toggle spam_reaction in this chat\n\n"),
	styling.Code(
		"!kandinsky",
	), styling.Plain(" - generate image via "), styling.TextURL("kandinsky", "https://www.sberbank.com/promo/kandinsky/"),
}

func (r *Presentation) helpCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	_, err := ctx.Reply(update, helpMessage, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
