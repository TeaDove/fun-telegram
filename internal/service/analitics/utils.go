package analitics

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strings"
)

type nameGetter struct {
	m map[int64]string
}

func (r *nameGetter) Get(userId int64) string {
	name, ok := r.m[userId]
	if !ok || strings.TrimSpace(name) == "" {
		return fmt.Sprintf("id: %d", userId)
	}

	return name
}

func (r *Service) getNameGetter(ctx context.Context, chatId int64) (nameGetter, error) {
	tgUsers, err := r.dbRepository.GetUsersByChatId(ctx, chatId)
	if err != nil {
		return nameGetter{}, errors.WithStack(err)
	}

	idToName := make(map[int64]string, len(tgUsers))
	for _, user := range tgUsers {
		idToName[user.TgUserId] = user.TgName
	}

	return nameGetter{m: idToName}, nil
}
