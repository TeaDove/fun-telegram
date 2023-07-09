package memory

import (
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/teadove/goteleout/internal/utils"
)

func saveLoad(t *testing.T) *Storage {
	storage := MustNew(true, "/tmp/test.json")

	err := storage.Save("key", []byte("value"))
	utils.Check(err)

	res, err := storage.Load("key")
	utils.Check(err)

	assert.Equal(t, []byte("value"), res)

	return storage
}

func TestUnit_MemoryStorage_SaveLoad_Ok(t *testing.T) {
	saveLoad(t)
}

func TestUnit_MemoryStorage_Flush_Ok(t *testing.T) {
	storage := saveLoad(t)

	err := storage.flush()
	utils.Check(err)
}

func TestUnit_MemoryStorage_loadFlushed_Ok(t *testing.T) {
	storage := saveLoad(t)
	err := storage.Save("abc", []byte("def"))
	utils.Check(err)
	oldMapping := storage.mapping

	err = storage.flush()
	utils.Check(err)

	err = storage.loadFlushed()
	utils.Check(err)

	assert.Equal(t, oldMapping, storage.mapping)
}
