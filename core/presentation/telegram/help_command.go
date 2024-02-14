package telegram

import (
	"context"
	"fmt"
	"slices"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/service/resource"
	"golang.org/x/exp/maps"
)

func (r *Presentation) compileHelpMessage(ctx context.Context, input *Input) []styling.StyledTextOption {
	helpMessage := make([]styling.StyledTextOption, 0, 20)
	helpMessage = append(
		helpMessage,
		styling.Plain(r.resourceService.Localize(ctx, resource.CommandHelpBegin, input.Locale)),
	)

	keys := maps.Keys(r.router)
	slices.Sort(keys)

	for _, commandName := range keys {
		command := r.router[commandName]
		helpMessage = append(
			helpMessage,
			styling.Plain(
				fmt.Sprintf(
					"/%s - %s\n",
					commandName,
					r.resourceService.Localize(ctx, command.description, input.Locale),
				),
			),
		)
		if command.requireAdmin {
			helpMessage = append(helpMessage,
				styling.Bold(r.resourceService.Localize(ctx, resource.AdminRequires, input.Locale)),
				styling.Plain("\n"),
			)
		}
		if command.requireOwner {
			helpMessage = append(helpMessage,
				styling.Bold(r.resourceService.Localize(ctx, resource.OwnerRequires, input.Locale)),
				styling.Plain("\n"),
			)
		}
		if command.example != "" {
			helpMessage = append(helpMessage,
				styling.Plain(r.resourceService.Localize(ctx, resource.Example, input.Locale)),
				styling.Plain(": "),
				styling.Code(
					fmt.Sprintf("!%s %s\n",
						commandName,
						command.example,
					)))
		}

		if len(command.flags) == 0 {
			helpMessage = append(helpMessage, styling.Plain("\n"))
			continue
		}

		for _, flag := range command.flags {
			helpMessage = append(
				helpMessage,
				styling.Plain(
					fmt.Sprintf(
						"%s/%s - %s\n",
						flag.Long,
						flag.Short,
						r.resourceService.Localize(ctx, flag.Description, input.Locale),
					),
				),
			)
		}
		helpMessage = append(helpMessage, styling.Plain("\n"))
	}

	return helpMessage
}

func (r *Presentation) helpCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	_, err := ctx.Reply(update, r.compileHelpMessage(ctx, input), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
