package analitics

import (
	"context"
	stderrors "errors"

	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/shared"
)

var ErrNotAnImage = errors.New("inputed filed is not an image")

func (r *Service) AnimePrediction(ctx context.Context, animeImage []byte) (float64, error) {
	resp, err := r.dsSupplier.AnimePrediction(ctx, animeImage)
	if err != nil {
		var dsError ds_supplier.DSError
		if errors.As(err, &dsError) && dsError.Detail == "inputed file is not an image" {
			return 0, stderrors.Join(ErrNotAnImage, err)
		}

		return 0, errors.Wrap(err, "failed to predict anime")
	}

	zerolog.Ctx(ctx).
		Debug().
		Str("status", "anime.predicted").
		Float64("conf", shared.ToFixed(resp, 3)).
		Send()

	return resp, nil
}
