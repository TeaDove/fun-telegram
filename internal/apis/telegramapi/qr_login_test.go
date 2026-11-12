package telegramapi

import (
	"strings"
	"testing"

	"github.com/celestix/gotgproto"
	protoerrors "github.com/celestix/gotgproto/errors"
	"github.com/cockroachdb/errors"

	"github.com/stretchr/testify/require"
	"rsc.io/qr"
)

func TestQRScreenRendersCodeWithoutPlaintextToken(t *testing.T) {
	t.Parallel()

	const fixture = "tg://login?token=fixture"

	code, err := qr.Encode(fixture, qr.M)
	require.NoError(t, err)

	screen := qrScreen(code)
	width := code.Size + 2*qrQuietZone

	require.NotContains(t, screen, fixture)
	require.Equal(t, width*((width+1)/2), strings.Count(screen, "▀"))
}

func TestConnectWithQR(t *testing.T) {
	t.Parallel()

	fixtureErr := errors.New("fixture failure")
	unauthorized := errors.Wrap(protoerrors.ErrSessionUnauthorized, "connect")
	testCases := []struct {
		name       string
		connectErr error
		loginErr   error
		wantErr    error
		wantCalls  []string
	}{
		{
			name:      "authorized session",
			wantCalls: []string{"connect"},
		},
		{
			name:       "new session",
			connectErr: unauthorized,
			wantCalls:  []string{"connect", "login", "connect"},
		},
		{
			name:       "connection failure",
			connectErr: fixtureErr,
			wantErr:    fixtureErr,
			wantCalls:  []string{"connect"},
		},
		{
			name:       "login failure",
			connectErr: unauthorized,
			loginErr:   fixtureErr,
			wantErr:    fixtureErr,
			wantCalls:  []string{"connect", "login"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := &gotgproto.Client{}
			calls := make([]string, 0, 3)
			connect := func() (*gotgproto.Client, error) {
				calls = append(calls, "connect")
				if len(calls) == 1 && testCase.connectErr != nil {
					return nil, testCase.connectErr
				}

				return client, nil
			}
			login := func() error {
				calls = append(calls, "login")

				return testCase.loginErr
			}

			result, err := connectWithQR(connect, login)
			require.ErrorIs(t, err, testCase.wantErr)
			require.Equal(t, testCase.wantCalls, calls)

			if testCase.wantErr == nil {
				require.Same(t, client, result)
			} else {
				require.Nil(t, result)
			}
		})
	}
}
