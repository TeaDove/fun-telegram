package redis_repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/fun_telegram/core/shared"
)

type User struct {
	Age     int    `redis:"age"`
	Name    string `redis:"name"`
	Deleted bool   `redis:"deleted"`
}

func TestIntegration_RedisStorage_HSet_Ok(t *testing.T) {
	t.Parallel()

	ctx := shared.GetCtx()
	storage := getStorage()
	hkey := shared.RandomString()

	err := storage.HSet(ctx, hkey, "age", 20)
	assert.NoError(t, err)

	err = storage.HSet(ctx, hkey, "name", shared.RandomString())
	assert.NoError(t, err)
}

func TestIntegration_RedisStorage_HGetAll_Ok(t *testing.T) {
	t.Parallel()

	ctx := shared.GetCtx()
	storage := getStorage()
	hkey := shared.RandomString()

	err := storage.HSet(ctx, hkey, "age", 20)
	require.NoError(t, err)

	name := shared.RandomString()
	err = storage.HSet(ctx, hkey, "name", name)
	require.NoError(t, err)

	err = storage.HSet(ctx, hkey, "deleted", true)
	require.NoError(t, err)

	user := User{}
	err = storage.HGetAll(ctx, hkey, &user)
	assert.NoError(t, err)
	assert.Equal(t, 20, user.Age)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, true, user.Deleted)
}

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

func TestIntegration_RedisStorage_HGetAll_NotFound(t *testing.T) {
	t.Parallel()
	ctx := shared.GetCtx()
	storage := getStorage()

	k := shared.RandomString()

	var user struct{}

	err := storage.HGetAll(ctx, k, &user)
	assert.ErrorIs(t, err, ErrKeyNotFound)
}
