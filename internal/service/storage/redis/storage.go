package redis

import (
	"context"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/service/storage"
)

type Storage struct {
	rbs redis.Client
}

func MustNew() *Storage {
	return &Storage{rbs: *redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB,
	})}
}

// TODO move add context
func (r *Storage) Load(k string) ([]byte, error) {
	cmd := r.rbs.Get(context.Background(), k)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return []byte{}, errors.WithStack(storage.KeyError)
		}
		return []byte{}, errors.WithStack(cmd.Err())
	}

	return []byte(cmd.Val()), nil
}

func (r *Storage) Save(k string, t []byte) error {
	ctx := context.Background()
	cmd := r.rbs.Set(ctx, k, t, 0)
	if cmd.Err() != nil {
		return errors.WithStack(cmd.Err())
	}

	return nil
}

func (r *Storage) Contains(k string) bool {
	cmd := r.rbs.Get(context.Background(), k)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return false
		}
		log.Error().
			Err(errors.WithStack(cmd.Err())).
			Str("status", "unable.execute.get.command").
			Str("key", k).
			Send()
		return false
	}

	return true
}

func (r *Storage) Delete(k string) error {
	ctx := context.Background()
	cmd := r.rbs.Del(ctx, k)

	if cmd.Err() != nil {
		return errors.WithStack(cmd.Err())
	}

	return nil
}
