package analitics

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/teadove/goteleout/core/repository/mongo_repository"
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

func (r *Service) getNameGetter(usersInChat mongo_repository.UsersInChat) nameGetter {
	idToName := make(map[int64]string, len(usersInChat))
	for _, user := range usersInChat {
		idToName[user.TgId] = fmt.Sprintf("%s", user.TgName)
	}

	return nameGetter{m: idToName}
}

func (r *Service) getNameGetterFromChatId(ctx context.Context, chatId int64) (nameGetter, error) {
	usersInChat, err := r.mongoRepository.GetUsersInChat(ctx, chatId)
	if err != nil {
		return nameGetter{}, errors.Wrap(err, "failed to get users in chat from mongo repository")
	}

	return r.getNameGetter(usersInChat), nil
}
