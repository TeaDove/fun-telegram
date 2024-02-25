package ch_repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/fun_telegram/core/shared"
)

func getRepository(t *testing.T) *Repository {
	r, err := New(shared.GetCtx())
	require.NoError(t, err)

	return r
}

func TestIntegration_ChRepository_Ping_Ok(t *testing.T) {
	t.Parallel()

	r := getRepository(t)

	err := r.Ping(shared.GetCtx())
	assert.NoError(t, err)
}

func TestIntegration_ChRepository_MessageFindInterlocutors_Ok(t *testing.T) {
	t.Parallel()

	r := getRepository(t)

	output, err := r.MessageFindInterlocutors(shared.GetCtx(), 1701683862, 418878871, 10, time.Minute*5)
	assert.NoError(t, err)
	shared.SendInterface(output)
}
