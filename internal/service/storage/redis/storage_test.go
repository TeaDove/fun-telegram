package redis

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strings"
	"testing"
)

func getStorage() *Storage {
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

	storage := getStorage()

	err := storage.Save(randomString(), []byte(randomString()))
	assert.NoError(t, err)
}

func TestIntegration_RedisStorage_Load_Ok(t *testing.T) {
	t.Parallel()

	storage := getStorage()

	k, v := randomString(), []byte(randomString())

	err := storage.Save(k, v)
	assert.NoError(t, err)

	newV, err := storage.Load(k)
	assert.NoError(t, err)

	assert.Equal(t, newV, v)
}

func TestIntegration_RedisStorage_Contains_Ok(t *testing.T) {
	t.Parallel()

	storage := getStorage()

	k := randomString()

	err := storage.Save(k, []byte(randomString()))
	assert.NoError(t, err)

	ok := storage.Contains(k)
	assert.True(t, ok)
}

func TestIntegration_RedisStorage_Delete_Ok(t *testing.T) {
	t.Parallel()

	storage := getStorage()

	k := randomString()

	err := storage.Save(k, []byte(randomString()))
	assert.NoError(t, err)

	err = storage.Delete(k)
	assert.NoError(t, err)

	ok := storage.Contains(k)
	assert.False(t, ok)
}
