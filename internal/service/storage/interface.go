package storage

import (
	"github.com/pkg/errors"
)

var KeyError = errors.New("key error")

type Interface interface {
	Load(k string) ([]byte, error)
	Save(k string, t []byte) error
	Contains(k string) bool
	Delete(k string) error
}
