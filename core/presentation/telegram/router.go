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
	executor    func(c *Context) error
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

	c := getOpt(text, route.flags...)
	c.extCtx = ctx
	c.update = update
	c.presentation = r

	if update.EffectiveUser().GetID() != ctx.Self.ID {
		_, err := ctx.Reply(update, ext.ReplyTextString("Err: insufficient privilege: owner rights required"), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	c.StartedAt = time.Now().UTC()

	zerolog.Ctx(ctx.Context).
		Debug().
		Str("command", firstWord).
		Msg("executing.command.begin")

	err := route.executor(&c)
	elapsed := time.Now().UTC().Sub(c.StartedAt)

	if err != nil {
		zerolog.Ctx(ctx.Context).Error().
			Stack().Err(errors.WithStack(err)).
			Str("elapsed", elapsed.String()).
			Msg("failed.to.process.command")

		errMessage := fmt.Sprintf("Err: something went wrong... : %e", err)

		var innerErr error

		if c.Silent {
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
