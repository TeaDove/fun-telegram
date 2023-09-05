package utils

import (
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
)

func GetNameFromPeerUser(user *peers.User) string {
	var returnName string
	name, ok := user.FirstName()
	if ok {
		returnName = name
	}
	name, ok = user.LastName()
	if ok {
		if returnName == "" {
			return name
		}
		returnName += " " + name
	}
	if returnName != "" {
		return returnName
	}

	name, ok = user.Username()
	if ok {
		return name
	}
	return "undefined"
}

func GetNameFromUser(user *tg.User) string {
	var returnName string
	name := user.FirstName
	if name != "" {
		returnName = name
	}
	name = user.LastName
	if name != "" {
		if returnName == "" {
			return name
		}
		returnName += " " + name
	}
	if returnName != "" {
		return returnName
	}

	name = user.Username
	if name != "" {
		return name
	}
	return "undefined"
}
