package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/service/storage"
	"strconv"
	"strings"
)

type messageProcessor struct {
	flags    []tgUtils.OptFlag
	executor func(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error
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

	chatId := strconv.Itoa(int(update.EffectiveChat().GetID()))

	_, err := r.storage.Load(chatId)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			// Bot enabled
		} else {
			return errors.WithStack(err)
		}
	} else {
		if command != "disable" {
			zerolog.Ctx(ctx.Context).
				Debug().
				Str("status", "bot.disable.in.chat").
				Str("chat_id", chatId).
				Str("command", command).
				Send()

			return nil
		}
		// only pass not disable command
	}

	zerolog.Ctx(ctx.Context).
		Info().
		Str("status", "executing.command").
		Interface("input", opts).
		Str("command", firstWord).
		Send()

	err = route.executor(ctx, update, &opts)
	if err != nil {
		log.Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.process.command").Send()
		return errors.WithStack(err)
	}

	return nil
}
