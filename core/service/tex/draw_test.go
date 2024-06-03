package tex

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnit_TexServic_Draw_Ok(t *testing.T) {
	f, err := os.Create(".test-output.png")
	require.NoError(t, err)
	defer f.Close()

	service := Service{}
	err = service.Draw(DrawInput{
		Write: f,
		Expr:  "Найс! $f(x) = \\frac{\\sqrt{x +20}}{2\\pi} +\\hbar \\sum y\\partial y$",
		Size:  50,
		DPI:   90,
	})

	require.NoError(t, err)
}
