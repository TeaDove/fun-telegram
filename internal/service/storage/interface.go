package storage

import (
	"github.com/pkg/errors"
)

var ErrKeyNotFound = errors.New("key not found")

type Interface interface {
	Load(k string) ([]byte, error)
	Save(k string, t []byte) error
	Contains(k string) bool
	Delete(k string) error
}
