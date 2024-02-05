package storage

import (
	"context"
	"github.com/pkg/errors"
)

var ErrKeyNotFound = errors.New("key not found")

type Interface interface {
	Load(k string) ([]byte, error)
	Save(k string, t []byte) error
	Contains(k string) bool
	Delete(k string) error
	// Toggle
	// Toggles k
	// Returns true, if k WAS toggled on
	Toggle(k string) (bool, error)
	GetToggle(k string) (bool, error)

	Ping(ctx context.Context) error
}
