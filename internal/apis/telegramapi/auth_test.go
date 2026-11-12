package telegramapi

import (
	"bytes"
	"testing"

	"github.com/celestix/gotgproto"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestAuthStatusLogsOnlyDeliveryType(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer

	logger := zerolog.New(&logs)
	conversation := authConversator{
		AuthConversator: gotgproto.BasicConversator(),
		logger:          &logger,
	}

	conversation.AuthStatus(gotgproto.AuthStatus{Event: gotgproto.AuthStatusPhoneAsked})
	require.Empty(t, logs.String())

	conversation.AuthStatus(gotgproto.AuthStatus{
		Event:        gotgproto.AuthStatusPhoneCodeAsked,
		SentCodeType: &tg.AuthSentCodeTypeEmailCode{EmailPattern: "[REDACTED TEST EMAIL]"},
	})
	require.JSONEq(
		t,
		`{"level":"info","delivery":"auth.sentCodeTypeEmailCode","message":"telegram.auth.code.delivery"}`,
		logs.String(),
	)
}
