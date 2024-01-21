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

func trimUnprintable(v string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, v)
}

func GetNameFromTgUser(user *tg.User) string {
	var result string

	name, ok := user.GetFirstName()
	if ok && strings.TrimSpace(name) != "" {
		lastName, ok := user.GetLastName()
		if ok {
			result = fmt.Sprintf("%s %s", name, lastName)
		} else {
			result = name
		}
	}

	result = trimUnprintable(result)

	if strings.TrimSpace(result) == "" {
		username, ok := user.GetUsername()
		if ok {
			result = username
		} else {
			result = utils.Undefined
		}
	}

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
