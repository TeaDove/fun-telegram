package telegram

import (
	"context"
	"fmt"
	"slices"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/service/resource"
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
			r.resourceService.Localize(ctx, resource.CommandHelpBegin, input.ChatSettings.Locale),
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
					"/%s - %s\n",
					commandName,
					r.resourceService.Localize(ctx, command.description, input.ChatSettings.Locale),
				),
			),
		)
		for _, flag := range command.flags {
			helpMessage = append(
				helpMessage,
				styling.Code(fmt.Sprintf("-%s", flag.Short)),
				styling.Plain("/"),
				styling.Code(fmt.Sprintf("--%s", flag.Long)),
				styling.Plain(
					fmt.Sprintf(
						" - %s\n",
						r.resourceService.Localize(
							ctx,
							flag.Description,
							input.ChatSettings.Locale,
						),
					),
				),
			)
		}

		if command.requireAdmin {
			helpMessage = append(
				helpMessage,
				styling.Bold(
					r.resourceService.Localize(
						ctx,
						resource.AdminRequires,
						input.ChatSettings.Locale,
					),
				),
				styling.Plain("\n"),
			)
		}

		if command.requireOwner {
			helpMessage = append(
				helpMessage,
				styling.Bold(
					r.resourceService.Localize(
						ctx,
						resource.OwnerRequires,
						input.ChatSettings.Locale,
					),
				),
				styling.Plain("\n"),
			)
		}

		if command.example != "" {
			helpMessage = append(
				helpMessage,
				styling.Plain(
					r.resourceService.Localize(ctx, resource.Example, input.ChatSettings.Locale),
				),
				styling.Plain(": "),
				styling.Code(
					fmt.Sprintf("!%s %s\n",
						commandName,
						command.example,
					)),
			)
		}

		if !r.checkFeatureEnabled(&input.ChatSettings, commandName) {
			helpMessage = append(
				helpMessage,
				styling.Bold(
					r.resourceService.Localize(
						ctx,
						resource.CommandHelpDisabled,
						input.ChatSettings.Locale,
					),
				),
				styling.Plain("\n"),
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
