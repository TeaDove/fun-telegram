package utils

import (
	"fmt"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/teadove/goteleout/internal/utils"
	"strings"
	"unicode"
)

func GetNameFromPeerUser(user *peers.User) string {
	tgUser := tg.User{}

	firstName, ok := user.FirstName()
	if ok {
		tgUser.SetFirstName(firstName)
	}

	lastName, ok := user.LastName()
	if ok {
		tgUser.SetLastName(lastName)
	}

	username, ok := user.Username()
	if ok {
		tgUser.SetUsername(username)
	}

	return GetNameFromTgUser(&tgUser)
}

func GetNameFromTgUser(user *tg.User) string {
	var result string

	name, nameOk := user.GetFirstName()
	lastName, lastNameOk := user.GetLastName()

	if nameOk {
		if lastNameOk {
			result = fmt.Sprintf("%s %s", name, lastName)
		}

		result = name
	}

	username, ok := user.GetUsername()
	if ok {
		result = username
	}

	if result == "" {
		result = utils.Undefined
	}

	result = strings.Map(func(r rune) rune {
		if unicode.IsGraphic(r) {
			return r
		}
		return -1
	}, result)

	return result
}

func GetChatName(chat types.EffectiveChat) string {
	switch v := chat.(type) {
	case *types.Channel:
		return v.Title
	case *types.Chat:
		return v.Title
	case *types.User:
		return v.Username
	}

	return utils.Undefined
}
