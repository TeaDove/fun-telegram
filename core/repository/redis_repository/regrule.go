package redis_repository

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
)

func getPathRegRule(chatId int64) string {
	return fmt.Sprintf("reg-rule::%d", chatId)
}

func (r *Repository) GetRegRules(ctx context.Context, chatId int64) (map[string]string, error) {
	cmd := r.rbs.HGetAll(ctx, getPathRegRule(chatId))
	map_, err := cmd.Result()
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "failed to hgetall")
	}

	return map_, nil
}

func (r *Repository) SetRegRules(ctx context.Context, chatId int64, k, v string) error {
	err := r.HSet(ctx, getPathRegRule(chatId), k, v)
	if err != nil {
		return errors.Wrap(err, "failed to hset")
	}

	return nil
}

func (r *Repository) DelRegRules(ctx context.Context, chatId int64, k string) error {
	cmd := r.rbs.HDel(ctx, getPathRegRule(chatId), k)
	if cmd.Err() != nil {
		return errors.Wrap(cmd.Err(), "failed to hdel")
	}

	return nil
}
