package mongo_repository

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	mgm.DefaultModel `bson:",inline"`

	TgChatID int64 `bson:"tg_chat_id"`
	TgId     int   `bson:"tg_id"`
	TgUserId int64 `bson:"tg_user_id"`
	Text     string
}

type RestartMessage struct {
	mgm.DefaultModel `bson:",inline"`

	MessageId primitive.ObjectID `bson:"message_id,omitempty"`
}

type User struct {
	mgm.DefaultModel `bson:",inline"`

	TgId       int64  `bson:"tg_id"`
	TgUsername string `bson:"tg_username"`
	TgName     string `bson:"tg_name"`
	IsBot      bool   `bson:"is_bot"`
}

type MemberStatus string

const (
	Plain MemberStatus = "PLAIN"
	// Creator is status for chat/channel creator.
	Creator MemberStatus = "CREATOR"
	// Admin is status for chat/channel admin.
	Admin MemberStatus = "ADMIN"
	// Banned is status for banned user.
	Banned MemberStatus = "BANNED"
	// Left is status for user that left chat/channel.
	Left MemberStatus = "LEFT"

	Unknown MemberStatus = "UNKNOWN"
)

type Member struct {
	mgm.DefaultModel `bson:",inline"`

	TgUserId int64 `bson:"tg_user_id"`
	TgChatId int64 `bson:"tg_chat_id"`
	Status   MemberStatus
}

type Chat struct {
	mgm.DefaultModel `bson:",inline"`

	TgId  int64 `bson:"tg_chat_id"`
	Title string
}

type UserInChat struct {
	TgId       int64  `bson:"tg_id"`
	TgUsername string `bson:"tg_username"`
	TgName     string `bson:"tg_name"`
	IsBot      bool   `bson:"is_bot"`
	Status     MemberStatus
}

type UsersInChat []UserInChat

func (r UsersInChat) ToMap() map[int64]UserInChat {
	map_ := make(map[int64]UserInChat, len(r))
	for _, user := range r {
		map_[user.TgId] = user
	}

	return map_
}
