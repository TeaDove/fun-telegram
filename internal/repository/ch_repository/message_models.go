package ch_repository

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	TgChatID int64  `json:"tg_chat_id"`
	TgUserId int64  `json:"tg_user_id"`
	Text     string `json:"text"`
	TgId     int64  `json:"tg_id"`
}
