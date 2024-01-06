package user_repository

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const userInChatCollectionName = "user_in_chat"

type UserInChat struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`

	TgChatID int64              `bson:"tg_chat_id"`
	UserID   primitive.ObjectID `bson:"user_id"`
	Toxicity float64            `bson:"toxicity"`
}

const userCollectionName = "user"

type User struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`

	TgUserID   int64  `bson:"tg_user_id"`
	TgUsername string `bson:"tg_username"`
}
