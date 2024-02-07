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

func (r *Presentation) setHelpMessage() {
	helpMessage := make([]styling.StyledTextOption, 0, 20)
	helpMessage = append(helpMessage,
		styling.Plain("Bot created by @TeaDove\nSource code: "),
		styling.TextURL("fun-telegram", "https://github.com/TeaDove/fun-telegram"),
		styling.Plain("\nAvailable commands:\n\n"),
	)

	keys := maps.Keys(r.router)
	slices.Sort(keys)

	for _, commandName := range keys {
		command := r.router[commandName]
		helpMessage = append(helpMessage, styling.Plain(fmt.Sprintf("/%s - %s\n", commandName, command.description)))
		if command.requireAdmin {
			helpMessage = append(helpMessage, styling.Bold("requires admin rights\n"))
		}
		if command.requireOwner {
			helpMessage = append(helpMessage, styling.Bold("requires owner rights\n"))
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
