package memory

import (
	"github.com/go-playground/assert/v2"
	"github.com/teadove/goteleout/internal/utils"
	"testing"
)

func TestUnit_MemoryStorage_SaveLoad_Ok(t *testing.T) {
	storage := MustNew()

	err := storage.Save("key", []byte("value"))
	utils.Check(err)

	res, err := storage.Load("key")
	utils.Check(err)

	assert.Equal(t, []byte("value"), res)
}
