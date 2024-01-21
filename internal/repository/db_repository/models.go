package db_repository

import (
	"github.com/kamva/mgm/v3"
)

type Message struct {
	mgm.DefaultModel `bson:",inline"`

	TgChatID int64 `bson:"tg_chat_id"`
	TgUserId int64 `bson:"tg_user_id"`
	Text     string
	TgId     int `bson:"tg_id"`
}

type User struct {
	mgm.DefaultModel `bson:",inline"`

	TgUserId   int64  `bson:"tg_user_id"`
	TgUsername string `bson:"tg_username"`
	TgName     string `bson:"tg_name"`
}
