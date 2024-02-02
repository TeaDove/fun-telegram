package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"strings"
	"time"
)

type messageProcessor struct {
	executor func(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error
	flags    []tgUtils.OptFlag
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

	opts := tgUtils.GetOpt(text, route.flags...)

	ok, err := r.isEnabled(update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	if !ok && command != "disable" {
		zerolog.Ctx(ctx.Context).
			Debug().
			Str("status", "bot.disable.in.chat").
			Str("command", command).
			Send()

		return nil
	}

	t0 := time.Now().UTC()
	zerolog.Ctx(ctx.Context).
		Info().
		Str("status", "executing.command.begin").
		Interface("input", opts).
		Str("command", firstWord).
		Send()

	err = route.executor(ctx, update, &opts)
	elapsed := time.Now().UTC().Sub(t0)

	if err != nil {
		zerolog.Ctx(ctx.Context).
			Error().
			Stack().
			Err(errors.WithStack(err)).
			Str("status", "failed.to.process.command").
			Dur("elapsed", elapsed).
			Send()
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx.Context).
		Info().
		Str("status", "executing.command.done").
		Dur("elapsed", elapsed).
		Send()

	return nil
}
