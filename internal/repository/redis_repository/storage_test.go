package redis_repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/teadove/goteleout/internal/shared"
	"math/rand"
	"strings"
	"testing"
)

func getStorage() *Repository {
	return MustNew()
}

func randomString() string {
	const alfabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var builder strings.Builder
	for i := 0; i < 50; i++ {
		builder.WriteByte(alfabet[rand.Intn(len(alfabet))])
	}

	return builder.String()
}

func TestIntegration_RedisStorage_Save_Ok(t *testing.T) {
	t.Parallel()
	ctx := shared.GetCtx()
	storage := getStorage()

	err := storage.Save(ctx, randomString(), []byte(randomString()))
	assert.NoError(t, err)
}

func TestIntegration_RedisStorage_Load_Ok(t *testing.T) {
	t.Parallel()
	ctx := shared.GetCtx()
	storage := getStorage()

	k, v := randomString(), []byte(randomString())

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

	k := randomString()

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
