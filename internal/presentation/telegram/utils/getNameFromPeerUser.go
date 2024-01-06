package utils

import (
	"fmt"
	"github.com/gotd/td/telegram/peers"
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

	return "undefined"
}
