package ch_repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/goteleout/internal/utils"
	"testing"
)

func getRepository(t *testing.T) *Repository {
	r, err := New(utils.GetCtx())
	require.NoError(t, err)

	return r
}

func TestIntegration_ChRepository_Ping_Ok(t *testing.T) {
	t.Parallel()

	r := getRepository(t)

	err := r.Ping(utils.GetCtx())
	assert.NoError(t, err)
}
