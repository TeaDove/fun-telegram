package messageservice

import (
	"fmt"
	"strings"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
)

type NameGetter struct {
	idToUser map[int64]tgdtos.Author
}

func (r *NameGetter) GetName(userID int64) string {
	user, ok := r.idToUser[userID]
	if !ok || strings.TrimSpace(user.Name) == "" {
		return fmt.Sprintf("id: %d", userID)
	}

	return user.Name
}

func (r *NameGetter) GetNameAndUsername(userID int64) string {
	user, ok := r.idToUser[userID]
	if !ok || (strings.TrimSpace(user.Name) == "" && strings.TrimSpace(user.Username) == "") {
		return fmt.Sprintf("id: %d", userID)
	}

	return fmt.Sprintf("%s (@%s)", user.Name, user.Username)
}

func (r UsersInChat) GetNameGetter() NameGetter {
	idToUser := make(map[int64]tgdtos.Author, len(r))

	for _, user := range r {
		idToUser[user.ID] = user
	}

	getter := NameGetter{idToUser: idToUser}

	return getter
}
