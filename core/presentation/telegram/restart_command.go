package telegram

import (
	"os"

	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func (r *Presentation) restartCommandHandler(c *Context) error {
	_, err := c.extCtx.SendMessage(c.extCtx.Self.ID, &tg.MessagesSendMessageRequest{Message: "Restarting..."})
	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	zerolog.Ctx(c.extCtx).Warn().Msg("reload.begin")

	os.Exit(0)

	return nil
}
