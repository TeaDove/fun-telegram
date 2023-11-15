package memory

import (
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/service/storage"
	"github.com/teadove/goteleout/internal/utils"
)

type Storage struct {
	mapping   map[string][]byte
	mappingMu *sync.RWMutex
	flushMu   *sync.Mutex

	persistent bool
	needFlush  bool
	filename   string
}

func MustNew(persistent bool, filename string) *Storage {
	mapping := make(map[string][]byte, 30)
	memoryStorage := Storage{mapping, &sync.RWMutex{}, &sync.Mutex{}, persistent, false, filename}

	scheduler := gocron.NewScheduler(time.UTC)

	if memoryStorage.persistent {
		err := memoryStorage.loadFlushed()
		utils.Check(err)
		_, err = scheduler.Every(1 * time.Minute).
			StartAt(time.Now().Add(1 * time.Minute)).
			Do(memoryStorage.flush)
		utils.Check(err)
		scheduler.StartAsync()
		log.Info().Str("status", "flush.scheduled").Str("filename", filename).Send()
	}

	return &memoryStorage
}

func (r *Storage) Load(k string) ([]byte, error) {
	r.mappingMu.RLock()
	defer r.mappingMu.RUnlock()

	v, ok := r.mapping[k]
	if !ok {
		return []byte{}, storage.ErrKeyNotFound
	}

	return v, nil
}

func (r *Storage) Save(k string, v []byte) error {
	r.mappingMu.Lock()
	defer r.mappingMu.Unlock()

	r.needFlush = true
	r.mapping[k] = v

	return nil
}

func (r *Storage) Delete(k string) error {
	r.mappingMu.Lock()
	defer r.mappingMu.Unlock()

	r.needFlush = true
	delete(r.mapping, k)

	return nil
}

func (r *Storage) Contains(k string) bool {
	r.mappingMu.Lock()
	defer r.mappingMu.Unlock()

	_, ok := r.mapping[k]

	return ok
}
