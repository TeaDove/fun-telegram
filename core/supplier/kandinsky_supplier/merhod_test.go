package kandinsky_supplier

import (
	"context"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/teadove/fun_telegram/core/shared"
)

var (
	ctx   = context.Background()
	input = RequestGenerationInput{
		Prompt: "Girl with green bear",
		Style:  "anime",
	}
)

func getSupplier(t *testing.T) *Supplier {
	supplier, err := New(ctx, shared.AppSettings.KandinskyKey, shared.AppSettings.KandinskySecret)
	assert.NoError(t, err)

	return supplier
}

func TestIntegration_KandinskySupplier_GetModels_Ok(t *testing.T) {
	t.Parallel()
	supplier := getSupplier(t)

	_, err := supplier.getModels(ctx)

	assert.NoError(t, err)
}

func TestIntegration_KandinskySupplier_RequestGeneration_Ok(t *testing.T) {
	t.Parallel()
	supplier := getSupplier(t)

	id, err := supplier.RequestGeneration(ctx, &input)

	assert.NoError(t, err)
	assert.NotNil(t, id)
}

func TestIntegration_KandinskySupplier_Get_Ok(t *testing.T) {
	t.Parallel()
	supplier := getSupplier(t)

	id, err := supplier.RequestGeneration(ctx, &input)
	assert.NoError(t, err)

	_, err = supplier.Get(ctx, id)
	assert.ErrorIs(t, err, ErrImageNotReady)
}

func TestIntegration_KandinskySupplier_WaitGet_Ok(t *testing.T) {
	t.Parallel()
	t.Skip("too long test")
	supplier := getSupplier(t)

	id, err := supplier.RequestGeneration(ctx, &input)
	assert.NoError(t, err)

	img, err := supplier.WaitGet(ctx, id)
	assert.NoError(t, err)
	assert.NotNil(t, img)
}

func TestIntegration_KandinskySupplier_WaitGeneration_Ok(t *testing.T) {
	t.Parallel()
	t.Skip("too long test")
	supplier := getSupplier(t)

	img, err := supplier.WaitGeneration(ctx, &input)
	assert.NoError(t, err)
	assert.NotNil(t, img)

	log.Info().Str("status", "image").Str("img", string(img)).Send()
}
