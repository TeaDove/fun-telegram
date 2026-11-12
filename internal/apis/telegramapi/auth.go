package telegramapi

import (
	"github.com/celestix/gotgproto"
	"github.com/rs/zerolog"
)

type authConversator struct {
	gotgproto.AuthConversator

	logger *zerolog.Logger
}

func (c authConversator) AuthStatus(status gotgproto.AuthStatus) {
	c.AuthConversator.AuthStatus(status)

	if status.SentCodeType == nil {
		return
	}

	c.logger.Info().
		Str("delivery", status.SentCodeType.TypeName()).
		Msg("telegram.auth.code.delivery")
}
