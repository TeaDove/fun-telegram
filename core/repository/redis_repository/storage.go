package redis_repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/teadove/fun_telegram/core/shared"
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

func (r *Repository) Load(ctx context.Context, k string) ([]byte, error) {
	cmd := r.rbs.Get(ctx, k)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			return []byte{}, errors.WithStack(ErrKeyNotFound)
		}

		return []byte{}, errors.WithStack(cmd.Err())
	}

	return []byte(cmd.Val()), nil
}

func (r *Repository) Save(ctx context.Context, k string, t []byte) error {
	err := r.rbs.Set(ctx, k, t, 0).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, k string) error {
	err := r.rbs.Del(ctx, k).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) Toggle(ctx context.Context, k string) (bool, error) {
	ok, err := r.GetToggle(ctx, k)
	if err != nil {
		return false, errors.WithStack(err)
	}

	if ok {
		err = r.Delete(ctx, k)
		if err != nil {
			return false, errors.WithStack(err)
		}

		return true, nil
	}

	err = r.Save(ctx, k, emptyBytes)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return false, nil
}

func (r *Repository) GetToggle(ctx context.Context, k string) (bool, error) {
	_, err := r.Load(ctx, k)
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
