package analitics

import (
	"fmt"
	"strings"

	"github.com/teadove/teasutils/utils/random_utils"

	"fun_telegram/core/repository/db_repository"
)

type nameGetter struct {
	idToUser     map[int64]db_repository.UserInChat
	anonymize    bool
	idToAnonName map[int64]string
}

func (r *nameGetter) contains(userID int64) bool {
	_, ok := r.idToUser[userID]
	return ok
}

func (r *nameGetter) getName(userID int64) string {
	if r.anonymize {
		return r.idToAnonName[userID]
	}

	user, ok := r.idToUser[userID]
	if !ok || strings.TrimSpace(user.TgName) == "" {
		return fmt.Sprintf("id: %d", userID)
	}

	return user.TgName
}

func (r *nameGetter) getNameAndUsername(userID int64) string {
	if r.anonymize {
		return r.idToAnonName[userID]
	}

	user, ok := r.idToUser[userID]
	if !ok || (strings.TrimSpace(user.TgName) == "" && strings.TrimSpace(user.TgUsername) == "") {
		return fmt.Sprintf("id: %d", userID)
	}

	return fmt.Sprintf("%s (@%s)", user.TgName, user.TgUsername)
}

func (r *Service) getNameGetter(
	usersInChat db_repository.UsersInChat,
	anonymize bool,
) nameGetter {
	idToUser := make(map[int64]db_repository.UserInChat, len(usersInChat))

	for _, user := range usersInChat {
		idToUser[user.TgID] = user
	}

	getter := nameGetter{idToUser: idToUser, anonymize: anonymize}

	if anonymize {
		getter.idToAnonName = make(map[int64]string, len(usersInChat))

		for _, user := range usersInChat {
			getter.idToAnonName[user.TgID] = random_utils.TextWithLen(6)
		}
	}

	return getter
}
