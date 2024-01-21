package toxicity

import (
	"context"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnit_ToxicityService_GetToxicity_Ok(t *testing.T) {
	r, err := New()
	assert.NoError(t, err)

	toxicity, err := r.GetToxicity(context.Background(), "заеб!")
	assert.NoError(t, err)
	assert.LessOrEqual(t, toxicity, 1.0)
	assert.GreaterOrEqual(t, toxicity, 0.0)

	log.Info().Str("status", "toxicity").Float64("toxicity", toxicity).Send()

	toxicity, err = r.GetToxicity(context.Background(), "привет!")
	assert.NoError(t, err)
	assert.LessOrEqual(t, toxicity, 1.0)
	assert.GreaterOrEqual(t, toxicity, 0.0)

	log.Info().Str("status", "toxicity").Float64("toxicity", toxicity).Send()
}
