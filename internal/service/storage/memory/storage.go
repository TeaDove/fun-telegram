package memory

import "github.com/teadove/goteleout/internal/service/storage"

type Storage struct {
	mapping map[string][]byte
}

func MustNew() *Storage {
	mapping := make(map[string][]byte, 30)
	memoryStorage := Storage{mapping}
	return &memoryStorage
}

func (r *Storage) Load(k string) ([]byte, error) {
	v, ok := r.mapping[k]
	if !ok {
		return []byte{}, storage.KeyError
	}
	return v, nil
}

func (r *Storage) Save(k string, v []byte) error {
	r.mapping[k] = v
	return nil
}
