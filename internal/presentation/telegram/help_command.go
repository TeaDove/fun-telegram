package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"golang.org/x/exp/maps"
	"slices"
)

var helpMessage = []styling.StyledTextOption{
	styling.Plain("Available commands:\n\n"),
	styling.Plain("/help"), styling.Plain(" - get this message\n\n"),
	styling.Plain("/ping"), styling.Plain(" - ping all users\n\n"),
	styling.Plain("/get_me"), styling.Plain(" - get id, username of requested user and group\n\n"),
	styling.Plain(
		"/spam_reaction [stop] [disable]",
	), styling.Plain(" - if replied to message with reaction, will spam this reaction to replied user.\n" +
		"[stop] - stop spamming.\n" +
		"[disable] - toggle spam_reaction in this chat\n\n"),
	styling.Plain(
		"/kandinsky",
	), styling.Plain(" - generate image via "), styling.TextURL("kandinsky", "https://www.sberbank.com/promo/kandinsky/"),
	styling.Plain("\n\n"),
	styling.Plain("/disable - fully disable bot in this chat\n\n"),
	styling.Plain("/location [address] - get description by ip address or domain\n\n"),
}

func (r *Presentation) setHelpMessage() {
	helpMessage = make([]styling.StyledTextOption, 0, 20)
	helpMessage = append(helpMessage, styling.Plain("Available commands:\n\n"))

	keys := maps.Keys(r.router)
	slices.Sort(keys)

	for _, commandName := range keys {
		command := r.router[commandName]
		helpMessage = append(helpMessage, styling.Plain(fmt.Sprintf("/%s - %s\n", commandName, command.description)))
		if command.requireAdmin {
			helpMessage = append(helpMessage, styling.Plain("requires admin rights\n"))
		}

		if len(command.flags) == 0 {
			helpMessage = append(helpMessage, styling.Plain("\n"))
			continue
		}

		for _, flag := range command.flags {
			helpMessage = append(
				helpMessage,
				styling.Plain(fmt.Sprintf("%s/%s - %s\n", flag.Long, flag.Short, flag.Description)),
			)
		}
		helpMessage = append(helpMessage, styling.Plain("\n"))
	}

	r.helpMessage = helpMessage
}

func (r *Presentation) helpCommandHandler(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error {
	_, err := ctx.Reply(update, r.helpMessage, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
