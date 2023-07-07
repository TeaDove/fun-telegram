package utils

import "github.com/gotd/td/tg"

func GetChatOrChannelId(entities *tg.Entities) int64 {
	for _, chat := range entities.Chats {
		return chat.GetID()
	}
	for _, chat := range entities.Channels {
		return chat.GetID()
	}
	return 0
}
