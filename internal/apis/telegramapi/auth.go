package telegramapi

import (
	"github.com/celestix/gotgproto"
	"github.com/rs/zerolog"
)

type authConversator struct {
	gotgproto.AuthConversator

	logger *zerolog.Logger
}

func (conversation authConversator) AuthStatus(status gotgproto.AuthStatus) {
	conversation.AuthConversator.AuthStatus(status)

	if status.SentCodeType == nil {
		return
	}

	conversation.logger.Info().
		Str("delivery", status.SentCodeType.TypeName()).
		Msg("telegram.auth.code.delivery")
}
