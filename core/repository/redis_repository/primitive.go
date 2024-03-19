package redis_repository

import (
	"context"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func (r *Repository) HSet(ctx context.Context, hkey string, values ...any) error {
	err := r.rbs.HSet(ctx, hkey, values...).Err()
	if err != nil {
		return errors.Wrap(err, "failed to hset")
	}

	zerolog.Ctx(ctx).Trace().Str("status", "redis.hset").Str("key", hkey).Send()

	return nil
}

func (r *Repository) HGetAll(ctx context.Context, key string, v any) error {
	cmd := r.rbs.HGetAll(ctx, key)
	map_, err := cmd.Result()

	if err != nil {
		return errors.Wrap(err, "failed to hgetall")
	}

	if len(map_) == 0 {
		return ErrKeyNotFound
	}

	err = cmd.Scan(v)
	if err != nil {
		return errors.Wrap(err, "failed to scan")
	}

	zerolog.Ctx(ctx).Trace().Str("status", "redis.hgetall").Str("key", key).Send()

	return nil
}

func (r *Repository) Load(ctx context.Context, k string) ([]byte, error) {
	cmd := r.rbs.Get(ctx, k)
	if cmd.Err() != nil {
		if errors.Is(cmd.Err(), redis.Nil) {
			zerolog.Ctx(ctx).Trace().Str("status", "redis.key.not.found").Str("key", k).Send()
			return []byte{}, errors.WithStack(ErrKeyNotFound)
		}

		return []byte{}, errors.WithStack(cmd.Err())
	}

	zerolog.Ctx(ctx).Trace().Str("status", "redis.loaded").Str("key", k).Send()

	return []byte(cmd.Val()), nil
}

func (r *Repository) Save(ctx context.Context, k string, v []byte) error {
	err := r.rbs.Set(ctx, k, v, 0).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx).Trace().Str("status", "redis.saved").Str("key", k).Send()

	return nil
}

func (r *Repository) Delete(ctx context.Context, k string) error {
	err := r.rbs.Del(ctx, k).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx).Trace().Str("status", "redis.deleted").Str("key", k).Send()

	return nil
}

// Toggle
// Return true, if toggle WAS true
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
