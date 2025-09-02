package message_service

import (
	"fmt"
	"strings"
)

type NameGetter struct {
	idToUser map[int64]UserInChat
}

func (r *NameGetter) GetName(userID int64) string {
	user, ok := r.idToUser[userID]
	if !ok || strings.TrimSpace(user.TgName) == "" {
		return fmt.Sprintf("id: %d", userID)
	}

	return user.TgName
}

func (r *NameGetter) GetNameAndUsername(userID int64) string {
	user, ok := r.idToUser[userID]
	if !ok || (strings.TrimSpace(user.TgName) == "" && strings.TrimSpace(user.TgUsername) == "") {
		return fmt.Sprintf("id: %d", userID)
	}

	return fmt.Sprintf("%s (@%s)", user.TgName, user.TgUsername)
}

func (r UsersInChat) GetNameGetter() NameGetter {
	idToUser := make(map[int64]UserInChat, len(r))

	for _, user := range r {
		idToUser[user.TgID] = user
	}

	getter := NameGetter{idToUser: idToUser}

	return getter
}
