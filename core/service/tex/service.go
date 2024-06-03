package tex

import (
	"io"

	"github.com/go-latex/latex/drawtex/drawimg"
	"github.com/go-latex/latex/mtex"
	"github.com/pkg/errors"
)

type Service struct{}

type DrawInput struct {
	Write io.Writer

	Expr string
	Size float64
	DPI  float64
}

func (r *Service) Draw(input DrawInput) error {
	err := mtex.Render(drawimg.NewRenderer(input.Write), input.Expr, input.Size, input.DPI, nil)
	if err != nil {
		return errors.Wrap(err, "failed to draw")
	}

	return nil
}
