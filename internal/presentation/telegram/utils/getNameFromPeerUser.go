package utils

import (
	"fmt"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/teadove/goteleout/internal/utils"
)

func GetNameFromPeerUser(user *peers.User) string {
	name, nameOk := user.FirstName()
	lastName, lastNameOk := user.LastName()

	if nameOk {
		if lastNameOk {
			return fmt.Sprintf("%s %s", name, lastName)
		}

		return name
	}

	username, ok := user.Username()
	if ok {
		return username
	}

	return utils.Undefined
}

func GetNameFromTgUser(user *tg.User) string {
	name, nameOk := user.GetFirstName()
	lastName, lastNameOk := user.GetLastName()

	if nameOk {
		if lastNameOk {
			return fmt.Sprintf("%s %s", name, lastName)
		}

		return name
	}

	username, ok := user.GetUsername()
	if ok {
		return username
	}

	return utils.Undefined
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
