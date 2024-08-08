package yt_supplier

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/fun_telegram/core/shared"
)

func TestIntegration_YtSupplier_GetVideo_Ok(t *testing.T) {
	ctx := shared.GetCtx()

	r, err := New(ctx)
	require.NoError(t, err)

	stream, err := r.GetVideo(ctx, "https://www.youtube.com/watch?v=dQw4w9WgXcQ")
	assert.NoError(t, err)

	file, err := os.Create("video.mp4")
	assert.NoError(t, err)
	defer file.Close()

	_, err = io.Copy(file, stream)
	assert.NoError(t, err)
}
