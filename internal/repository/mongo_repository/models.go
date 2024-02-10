package mongo_repository

import (
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	mgm.DefaultModel `bson:",inline"`

	TgChatID int64 `bson:"tg_chat_id"`
	TgUserId int64 `bson:"tg_user_id"`
	Text     string
	TgId     int `bson:"tg_id"`
}

type MessageStorageStats struct {
	TotalSizeBytes           int
	Count                    int
	AvgObjWithIndexSizeBytes int
}

type User struct {
	mgm.DefaultModel `bson:",inline"`

	TgUserId   int64  `bson:"tg_user_id"`
	TgUsername string `bson:"tg_username"`
	TgName     string `bson:"tg_name"`
	IsBot      bool   `bson:"is_bot"`
}

type RestartMessage struct {
	mgm.DefaultModel `bson:",inline"`

	MessageId primitive.ObjectID `bson:"message_id,omitempty"`
}
