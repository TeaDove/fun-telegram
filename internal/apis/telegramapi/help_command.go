package telegramapi

import (
	"fmt"
	"slices"

	"github.com/celestix/gotgproto/ext"
	"github.com/cockroachdb/errors"
	"github.com/gotd/td/telegram/message/styling"
	"golang.org/x/exp/maps"
)

func (r *Presentation) compileHelpMessage() []styling.StyledTextOption {
	helpMessage := make([]styling.StyledTextOption, 0, 20)
	helpMessage = append(
		helpMessage,
		styling.Plain(
			"Source code: https://github.com/TeaDove/fun-telegram\nAvailable commands:\n\n",
		),
	)

	keys := maps.Keys(r.router)
	slices.Sort(keys)

	for _, commandName := range keys {
		command := r.router[commandName]
		helpMessage = append(
			helpMessage,
			styling.Plain(
				fmt.Sprintf(
					"/%s - %s\n", commandName, command.description,
				),
			),
		)

		for _, flag := range command.flags {
			helpMessage = append(
				helpMessage,
				styling.Code(fmt.Sprintf("-%s", flag.Short)),
				styling.Plain("/"),
				styling.Code(fmt.Sprintf("--%s", flag.Long)),
				styling.Plain(fmt.Sprintf(" - %s\n", flag.Description)),
			)
		}

		if command.example != "" {
			helpMessage = append(
				helpMessage,
				styling.Plain("Example: "),
				styling.Code(
					fmt.Sprintf("!%s %s\n",
						commandName,
						command.example,
					)),
			)
		}

		helpMessage = append(helpMessage, styling.Plain("\n"))
	}

	return helpMessage
}

func (r *Presentation) helpCommandHandler(c *Context) error {
	_, err := c.extCtx.Reply(c.update, ext.ReplyTextStyledTextArray(r.compileHelpMessage()), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
