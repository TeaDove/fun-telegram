package redis_repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teadove/goteleout/core/shared"
)

func getStorage() *Repository {
	return MustNew()
}

func TestIntegration_RedisStorage_Save_Ok(t *testing.T) {
	t.Parallel()
	ctx := shared.GetCtx()
	storage := getStorage()

	err := storage.Save(ctx, shared.RandomString(), []byte(shared.RandomString()))
	assert.NoError(t, err)
}

func TestIntegration_RedisStorage_Load_Ok(t *testing.T) {
	t.Parallel()
	ctx := shared.GetCtx()
	storage := getStorage()

	k, v := shared.RandomString(), []byte(shared.RandomString())

	err := storage.Save(ctx, k, v)
	assert.NoError(t, err)

	newV, err := storage.Load(ctx, k)
	assert.NoError(t, err)

	assert.Equal(t, newV, v)
}

func TestIntegration_RedisStorage_Toggle_Ok(t *testing.T) {
	t.Parallel()
	ctx := shared.GetCtx()
	storage := getStorage()

	k := shared.RandomString()

	ok, err := storage.GetToggle(ctx, k)
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = storage.Toggle(ctx, k)
	assert.NoError(t, err)
	assert.False(t, ok)

	ok, err = storage.GetToggle(ctx, k)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = storage.Toggle(ctx, k)
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = storage.GetToggle(ctx, k)
	assert.NoError(t, err)
	assert.False(t, ok)
}
