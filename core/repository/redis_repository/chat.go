package redis_repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type ChatSettings struct {
	Enabled  bool   `redis:"enabled"`
	Locale   string `redis:"locale"`
	Tz       int8   `redis:"tz"`
	Features []byte `redis:"features"`
}

func getPathChatSettings(chatId int64) string {
	return fmt.Sprintf("chat-settings::%d", chatId)
}

func (r *Repository) GetChatSettings(ctx context.Context, chatId int64) (ChatSettings, error) {
	var chatSettings ChatSettings

	err := r.HGetAll(ctx, getPathChatSettings(chatId), &chatSettings)
	if err != nil {
		return ChatSettings{}, errors.Wrap(err, "failed to hgetall")
	}

	return chatSettings, nil
}

func (r *Repository) SetChatSettings(ctx context.Context, chatId int64, v *ChatSettings) error {
	err := r.HSet(ctx, getPathChatSettings(chatId), v)
	if err != nil {
		return errors.Wrap(err, "failed to hset")
	}

	return nil
}
