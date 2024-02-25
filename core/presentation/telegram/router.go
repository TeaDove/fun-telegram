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
	executor     func(ctx *ext.Context, update *ext.Update, input *Input) error
	description  resource.Code
	requireAdmin bool
	requireOwner bool
	flags        []OptFlag
	example      string
}

func (r *Presentation) route(ctx *ext.Context, update *ext.Update) error {
	_, ok := update.UpdateClass.(*tg.UpdateNewChannelMessage)
	if !ok {
		_, ok = update.UpdateClass.(*tg.UpdateNewMessage)
		if !ok {
			return nil
		}
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

	opts := GetOpt(text, route.flags...)

	ok, err := r.isEnabled(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.Wrap(err, "failed to check if enabled")
	}

	if !ok && command != "disable" {
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

	locale, err := r.getLocale(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	opts.Locale = locale

	if ok {
		_, err = ctx.Reply(update, r.resourceService.Localize(ctx, resource.ErrAccessDenies, locale), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	if route.requireAdmin {
		ok, err = r.checkFromAdmin(ctx, update)
		if err != nil {
			return errors.WithStack(err)
		}
		if !ok {
			_, err = ctx.Reply(
				update,
				r.resourceService.Localize(ctx, resource.ErrInsufficientPrivilegesAdmin, locale),
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
				r.resourceService.Localize(ctx, resource.ErrInsufficientPrivilegesOwner, locale),
				nil,
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}
	}

	opts.StartedAt = time.Now().UTC()
	zerolog.Ctx(ctx.Context).
		Info().
		Str("status", "executing.command.begin").
		Interface("input", opts).
		Str("command", firstWord).
		Send()

	err = route.executor(ctx, update, &opts)
	elapsed := time.Now().UTC().Sub(opts.StartedAt)

	if err != nil {
		zerolog.Ctx(ctx.Context).
			Error().
			Stack().
			Err(errors.WithStack(err)).
			Str("status", "failed.to.process.command").
			Dur("elapsed", elapsed).
			Send()

		errMessage := r.resourceService.Localizef(ctx, resource.ErrISE, opts.Locale, err.Error())
		var innerErr error

		if opts.Silent {
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
