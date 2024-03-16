package telegram

import (
	"strings"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/service/resource"
)

type messageProcessor struct {
	executor          func(ctx *ext.Context, update *ext.Update, input *input) error
	description       resource.Code
	requireAdmin      bool
	requireOwner      bool
	flags             []optFlag
	example           string
	disabledByDefault bool
}

// route
// nolint: gocyclo
func (r *Presentation) route(ctx *ext.Context, update *ext.Update) error {
	ok := filterNonNewMessages(update)
	if !ok {
		return nil
	}

	text := update.EffectiveMessage.Message.Message
	if len(text) == 0 || !(text[0] == '!' || text[0] == '/') {
		return nil
	}

	firstWord, _, _ := strings.Cut(text, " ")
	command := firstWord[1:]

	route, ok := r.router[command]
	if !ok {
		return nil
	}

	ctx.Context = zerolog.Ctx(ctx).With().Str("command", command).Ctx(ctx.Context).Logger().WithContext(ctx.Context)

	commandInput := GetOpt(text, route.flags...)

	chatSettings, err := r.getChatSettings(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.Wrap(err, "failed to get chat settings")
	}

	commandInput.ChatSettings = chatSettings

	if !chatSettings.Enabled && command != "chat" {
		zerolog.Ctx(ctx.Context).
			Debug().
			Str("status", "bot.disable.in.chat").
			Str("command", command).
			Send()

		return nil
	}

	ok, err = r.isBanned(ctx, update.EffectiveUser().Username)
	if err != nil {
		return errors.WithStack(err)
	}

	if ok {
		_, err = ctx.Reply(update, r.resourceService.Localize(ctx, resource.ErrAccessDenies, chatSettings.Locale), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	if !r.checkFeatureEnabled(&chatSettings, command) {
		_, err = ctx.Reply(update, r.resourceService.Localize(ctx, resource.ErrFeatureDisabled, chatSettings.Locale), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	if route.requireAdmin {
		ok, err = r.checkFromAdmin(ctx, update)
		if err != nil {
			return errors.Wrap(err, "failed to check if admin")
		}
		if !ok {
			_, err = ctx.Reply(
				update,
				r.resourceService.Localize(ctx, resource.ErrInsufficientPrivilegesAdmin, chatSettings.Locale),
				nil,
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}
	}
	if route.requireOwner {
		ok = r.checkFromOwner(ctx, update)
		if !ok {
			_, err = ctx.Reply(
				update,
				r.resourceService.Localize(ctx, resource.ErrInsufficientPrivilegesOwner, chatSettings.Locale),
				nil,
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}
	}

	commandInput.StartedAt = time.Now().UTC()
	zerolog.Ctx(ctx.Context).
		Info().
		Str("status", "executing.command.begin").
		//Interface("input", commandInput).
		Str("command", firstWord).
		Send()

	err = route.executor(ctx, update, &commandInput)
	elapsed := time.Now().UTC().Sub(commandInput.StartedAt)

	if err != nil {
		zerolog.Ctx(ctx.Context).
			Error().
			Stack().
			Err(errors.WithStack(err)).
			Str("status", "failed.to.process.command").
			Dur("elapsed", elapsed).
			Send()

		errMessage := r.resourceService.Localizef(ctx, resource.ErrISE, chatSettings.Locale, err.Error())
		var innerErr error

		if commandInput.Silent {
			_, innerErr = ctx.SendMessage(ctx.Self.ID, &tg.MessagesSendMessageRequest{Message: errMessage})
		} else {
			_, innerErr = ctx.Reply(update, errMessage, nil)
		}

		if innerErr != nil {
			zerolog.Ctx(ctx).Error().
				Stack().
				Err(err).
				Str("status", "failed.to.reply").
				Send()
			return nil
		}

		return nil
	}

	zerolog.Ctx(ctx.Context).
		Info().
		Str("status", "executing.command.done").
		Dur("elapsed", elapsed).
		Send()

	return nil
}
