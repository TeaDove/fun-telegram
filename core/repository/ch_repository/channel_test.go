package ch_repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/fun_telegram/core/shared"
)

func TestIntegration_ChRepository_ChannelSelect_NotFound(t *testing.T) {
	t.Parallel()

	r := getRepository(t)
	ctx := shared.GetCtx()

	_, err := r.ChannelSelect(ctx, 999)

	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestIntegration_ChRepository_ChannelSelect_Found(t *testing.T) {
	t.Parallel()

	r := getRepository(t)
	ctx := shared.GetCtx()

	err := r.ChannelInsert(ctx, &Channel{TgId: 12})
	require.NoError(t, err)

	channel, err := r.ChannelSelect(ctx, 12)
	assert.NoError(t, err)
	assert.Equal(t, int64(12), channel.TgId)
}

func TestIntegration_ChRepository_ChannelBatchInsert_Ok(t *testing.T) {
	t.Parallel()

	r := getRepository(t)
	ctx := shared.GetCtx()

	err := r.ChannelBatchInsert(ctx, []Channel{{TgId: 12, UploadedAt: time.Now()}, {TgId: 13, UploadedAt: time.Now()}})
	assert.NoError(t, err)
}
