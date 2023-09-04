package utils

import "github.com/gotd/td/telegram/peers"

func GetNameFromPeerUser(user *peers.User) string {
	name, ok := user.FirstName()
	if ok {
		return name
	}
	name, ok = user.LastName()
	if ok {
		return name
	}
	name, ok = user.Username()
	if ok {
		return name
	}
	return "undefined"
}
