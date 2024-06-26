package mongo_repository

import (
	"time"

	"github.com/gotd/td/tg"
	"github.com/teadove/fun_telegram/core/supplier/kandinsky_supplier"

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

	MessageId primitive.ObjectID `bson:"message_id"`
}

type PingMessage struct {
	mgm.DefaultModel `bson:",inline"`

	MessageId primitive.ObjectID `bson:"message_id"`
	DeleteAt  time.Time          `bson:"delete_at"`
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

func (r UsersInChat) ToIds() []int64 {
	slice := make([]int64, len(r))
	for _, user := range r {
		slice = append(slice, user.TgId)
	}

	return slice
}

type Image struct {
	mgm.DefaultModel `bson:",inline"`

	Content []byte `bson:"content"`
}

type TgImage struct {
	mgm.DefaultModel `bson:",inline"`

	TgInputPhoto tg.InputPhoto `bson:"tg_input_photo"`

	MessageId primitive.ObjectID `bson:"message_id"`
	ImageId   primitive.ObjectID `bson:"image_id"`
}

type KandinskyImage struct {
	mgm.DefaultModel `bson:",inline"`

	Input kandinsky_supplier.RequestGenerationInput `bson:"input"`

	TgImageId primitive.ObjectID `bson:"tg_image_id"`
	ImageId   primitive.ObjectID `bson:"image_id"`
}
