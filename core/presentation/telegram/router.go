package telegram

import (
	"fmt"
	"strings"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type messageProcessor struct {
	executor    func(ctx *ext.Context, update *ext.Update, input *input) error
	description string
	flags       []optFlag
	example     string
}

// route
// nolint: gocyclo // don't care
func (r *Presentation) route(ctx *ext.Context, update *ext.Update) error {
	text := update.EffectiveMessage.Message.Message
	if len(text) == 0 || (text[0] != '!' && text[0] != '/') {
		return nil
	}

	firstWord, _, _ := strings.Cut(text, " ")
	command := firstWord[1:]

	route, ok := r.router[command]
	if !ok {
		return nil
	}

	ctx.Context = zerolog.Ctx(ctx).
		With().
		Str("command", command).
		Ctx(ctx.Context).
		Logger().
		WithContext(ctx.Context)

	commandInput := getOpt(text, route.flags...)

	if update.EffectiveUser().GetID() != ctx.Self.ID {
		_, err := ctx.Reply(update, ext.ReplyTextString("Err: insufficient privilege: owner rights required"), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	commandInput.StartedAt = time.Now().UTC()

	zerolog.Ctx(ctx.Context).
		Info().
		// Interface("input", commandInput).
		Str("command", firstWord).
		Msg("executing.command.begin")

	err := route.executor(ctx, update, &commandInput)
	elapsed := time.Now().UTC().Sub(commandInput.StartedAt)

	if err != nil {
		zerolog.Ctx(ctx.Context).Error().
			Stack().Err(errors.WithStack(err)).
			Str("elapsed", elapsed.String()).
			Msg("failed.to.process.command")

		errMessage := fmt.Sprintf("Err: something went wrong... : %e", err)

		var innerErr error

		if commandInput.Silent {
			_, innerErr = ctx.SendMessage(
				ctx.Self.ID,
				&tg.MessagesSendMessageRequest{Message: errMessage},
			)
		} else {
			_, innerErr = ctx.Reply(update, ext.ReplyTextString(errMessage), nil)
		}

		if innerErr != nil {
			zerolog.Ctx(ctx).Error().
				Stack().
				Err(err).
				Msg("failed.to.reply")

			return nil // nolint: nilerr // as expected
		}

		return nil
	}

	zerolog.Ctx(ctx.Context).
		Info().
		Str("elapsed", elapsed.String()).
		Msg("executing.command.done")

	return nil
}
