package analitics

import (
	"fmt"
	"strings"

	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
)

type nameGetter struct {
	idToUser map[int64]mongo_repository.UserInChat
}

func (r *nameGetter) GetName(userId int64) string {
	user, ok := r.idToUser[userId]
	if !ok || strings.TrimSpace(user.TgName) == "" {
		return fmt.Sprintf("id: %d", userId)
	}

	return user.TgName
}

func (r *nameGetter) GetNameAndUsername(userId int64) string {
	user, ok := r.idToUser[userId]
	if !ok || (strings.TrimSpace(user.TgName) == "" && strings.TrimSpace(user.TgUsername) == "") {
		return fmt.Sprintf("id: %d", userId)
	}

	return fmt.Sprintf("%s (@%s)", user.TgName, user.TgUsername)
}

func (r *Service) getNameGetter(usersInChat mongo_repository.UsersInChat) nameGetter {
	idToUser := make(map[int64]mongo_repository.UserInChat, len(usersInChat))

	for _, user := range usersInChat {
		idToUser[user.TgId] = user
	}

	return nameGetter{idToUser: idToUser}
}
