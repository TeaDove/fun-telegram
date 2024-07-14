package analitics

import (
	"fmt"
	"strings"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/teadove/fun_telegram/core/shared"
)

type nameGetter struct {
	idToUser     map[int64]db_repository.UserInChat
	anonymize    bool
	idToAnonName map[int64]string
}

func (r *nameGetter) contains(userId int64) bool {
	_, ok := r.idToUser[userId]
	return ok
}

func (r *nameGetter) getName(userId int64) string {
	if r.anonymize {
		return r.idToAnonName[userId]
	}

	user, ok := r.idToUser[userId]
	if !ok || strings.TrimSpace(user.TgName) == "" {
		return fmt.Sprintf("id: %d", userId)
	}

	return user.TgName
}

func (r *nameGetter) getNameAndUsername(userId int64) string {
	if r.anonymize {
		return r.idToAnonName[userId]
	}

	user, ok := r.idToUser[userId]
	if !ok || (strings.TrimSpace(user.TgName) == "" && strings.TrimSpace(user.TgUsername) == "") {
		return fmt.Sprintf("id: %d", userId)
	}

	return fmt.Sprintf("%s (@%s)", user.TgName, user.TgUsername)
}

func (r *Service) getNameGetter(
	usersInChat db_repository.UsersInChat,
	anonymize bool,
) nameGetter {
	idToUser := make(map[int64]db_repository.UserInChat, len(usersInChat))

	for _, user := range usersInChat {
		idToUser[user.TgId] = user
	}

	getter := nameGetter{idToUser: idToUser, anonymize: anonymize}

	if anonymize {
		getter.idToAnonName = make(map[int64]string, len(usersInChat))

		for _, user := range usersInChat {
			getter.idToAnonName[user.TgId] = shared.RandomStringWithLength(6)
		}
	}

	return getter
}
