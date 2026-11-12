package telegramapi

import (
	"testing"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestCommandWithoutSenderRequiresSavedMessages(t *testing.T) {
	t.Parallel()

	c := Context{
		extCtx: &ext.Context{Self: &tg.User{ID: 1}},
		update: &ext.Update{EffectiveMessage: types.ConstructMessage(&tg.Message{
			PeerID: &tg.PeerUser{UserID: 2},
		})},
	}
	require.False(t, c.isOwnCommandMessage())

	msg := c.update.EffectiveMessage.Message
	msg.PeerID = &tg.PeerUser{UserID: 1}

	require.True(t, c.isOwnCommandMessage())

	msg.SetFwdFrom(tg.MessageFwdHeader{Date: 1})
	require.False(t, c.isOwnCommandMessage())
}
