package redis_repository

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/teadove/goteleout/internal/shared"
)

var emptyBytes = []byte{}
var ErrKeyNotFound = errors.New("key not found")

type Repository struct {
	rbs redis.Client
}

func MustNew() *Repository {
	return &Repository{rbs: *redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", shared.AppSettings.Storage.RedisHost),
		Password: "", // no password set
		DB:       0,  // use default DB,
	})}
}

func (r *Repository) Load(k string) ([]byte, error) {
	cmd := r.rbs.Get(context.Background(), k)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return []byte{}, errors.WithStack(ErrKeyNotFound)
		}

		return []byte{}, errors.WithStack(cmd.Err())
	}

	return []byte(cmd.Val()), nil
}

func (r *Repository) Save(k string, t []byte) error {
	ctx := context.Background()

	cmd := r.rbs.Set(ctx, k, t, 0)
	if cmd.Err() != nil {
		return errors.WithStack(cmd.Err())
	}

	return nil
}

func (r *Repository) Contains(k string) bool {
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

func (r *Repository) Delete(k string) error {
	ctx := context.Background()
	cmd := r.rbs.Del(ctx, k)

	if cmd.Err() != nil {
		return errors.WithStack(cmd.Err())
	}

	return nil
}

func (r *Repository) Toggle(k string) (bool, error) {
	_, err := r.Load(k)
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return false, errors.WithStack(err)
		}

		err = r.Save(k, emptyBytes)
		if err != nil {
			return false, errors.WithStack(err)
		}

		return false, nil
	}

	err = r.Delete(k)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return true, nil
}

func (r *Repository) GetToggle(k string) (bool, error) {
	_, err := r.Load(k)
	if err != nil {
		if !errors.Is(err, ErrKeyNotFound) {
			return false, errors.WithStack(err)
		}

		return false, nil
	}

	return true, nil
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.rbs.Ping(ctx).Err()
}
