package telegram

import (
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"strings"
)

type Input struct {
	Args   map[string]string
	Silent bool
}

type messageProcessor func(ctx *ext.Context, update *ext.Update, input *Input) error

func (r *Presentation) route(ctx *ext.Context, update *ext.Update) error {
	text := update.EffectiveMessage.Message.Message
	if len(text) == 0 || text[0] != '!' {
		return nil
	}

	command, _, _ := strings.Cut(text, " ")

	executor, ok := r.router[command[1:]]
	if !ok {
		return nil
	}

	args := tgUtils.GetArguments(text)
	_, silent := args["Silent"]

	input := Input{
		Args:   args,
		Silent: silent,
	}
	log.Debug().Str("status", "executing.command").Interface("input", input).Str("command", command).Send()

	err := executor(ctx, update, &input)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
