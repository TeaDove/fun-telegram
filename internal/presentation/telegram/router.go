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

type Input struct {
	Args   map[string]string
	Silent bool
}

type messageProcessor func(ctx *ext.Context, update *ext.Update, input *Input) error

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

	command, _, _ := strings.Cut(text, " ")

	executor, ok := r.router[command[1:]]
	if !ok {
		return nil
	}

	args := tgUtils.GetArguments(text)
	_, silent := args["silent"]

	input := Input{
		Args:   args,
		Silent: silent,
	}

	chatId := strconv.Itoa(int(update.EffectiveChat().GetID()))

	_, err := r.storage.Load(chatId)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			// Bot enabled
		} else {
			return errors.WithStack(err)
		}
	} else {
		if command[1:] != "disable" {
			zerolog.Ctx(ctx.Context).
				Debug().
				Str("status", "bot.disable.in.chat").
				Str("chat_id", chatId).
				Str("command", command[1:]).
				Send()

			return nil
		}
		// only pass not disable command
	}

	zerolog.Ctx(ctx.Context).
		Info().
		Str("status", "executing.command").
		Interface("input", input).
		Str("command", command).
		Send()

	err = executor(ctx, update, &input)
	if err != nil {
		log.Error().Stack().Err(errors.WithStack(err)).Str("status", "failed.to.process.command").Send()
		return errors.WithStack(err)
	}

	return nil
}
