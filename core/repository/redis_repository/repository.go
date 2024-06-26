package redis_repository

import (
	"fmt"

	"github.com/bsm/redislock"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/teadove/fun_telegram/core/shared"
)

var (
	emptyBytes     = []byte{}
	ErrKeyNotFound = errors.New("key not found")
)

type Repository struct {
	rbs    redis.Client
	Locker *redislock.Client
}

func MustNew() *Repository {
	client := *redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", shared.AppSettings.Storage.RedisHost),
		Password: "", // no password set
		DB:       0,  // use default DB,
	})
	locker := redislock.New(client)

	return &Repository{rbs: client, Locker: locker}
}
