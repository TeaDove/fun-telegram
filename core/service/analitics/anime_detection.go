package analitics

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/shared"
)

func (r *Service) AnimePrediction(ctx context.Context, animeImage []byte) (float64, error) {
	resp, err := r.dsSupplier.AnimePrediction(ctx, animeImage)
	if err != nil {
		return 0, errors.Wrap(err, "failed to predict anime")
	}

	zerolog.Ctx(ctx).
		Debug().
		Str("status", "anime.predicted").
		Float64("conf", shared.ToFixed(resp, 3)).
		Send()

	return resp, nil
}
