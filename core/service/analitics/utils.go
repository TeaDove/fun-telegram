package analitics

import (
	"fmt"
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
		idToName[user.TgId] = fmt.Sprintf("%s: @%s", user.TgName, user.TgUsername)
	}

	return nameGetter{m: idToName}
}
