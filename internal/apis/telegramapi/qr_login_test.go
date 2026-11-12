package telegramapi

import (
	"strings"
	"testing"

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
