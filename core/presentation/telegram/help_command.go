package telegram

import (
	"context"
	"fmt"
	"slices"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

func (r *Presentation) compileHelpMessage(
	ctx context.Context,
	input *input,
) []styling.StyledTextOption {
	helpMessage := make([]styling.StyledTextOption, 0, 20)
	helpMessage = append(
		helpMessage,
		styling.Plain(
			"Bot created by @TeaDove\nSource code: https://github.com/TeaDove/fun-telegram\nAvailable commands:\n\n",
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

func (r *Presentation) helpCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	_, err := ctx.Reply(update, r.compileHelpMessage(ctx, input), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
