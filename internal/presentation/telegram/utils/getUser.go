package utils

import "github.com/gotd/td/tg"

func GetPeer(entities *tg.Entities) tg.InputPeerClass {
	var peer tg.InputPeerClass
	for _, value := range entities.Channels {
		peer = value.AsInputPeer()
		return peer
	}
	for _, value := range entities.Chats {
		peer = value.AsInputPeer()
		return peer
	}
	for _, value := range entities.Users {
		peer = value.AsInputPeer()
		return peer
	}

	return peer
}
