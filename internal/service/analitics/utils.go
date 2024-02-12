package analitics

import (
	"fmt"
	"github.com/teadove/goteleout/internal/repository/mongo_repository"
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

func (r *Service) getNameGetter(usersInChat mongo_repository.UsersInChat) nameGetter {
	idToName := make(map[int64]string, len(usersInChat))
	for _, user := range usersInChat {
		idToName[user.TgId] = user.TgName
	}

	return nameGetter{m: idToName}
}
