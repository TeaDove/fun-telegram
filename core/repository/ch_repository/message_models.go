package ch_repository

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	Id        uuid.UUID `json:"id" ch:"id"`
	CreatedAt time.Time `json:"created_at" ch:"created_at"`

	TgChatID int64  `json:"tg_chat_id" ch:"tg_chat_id"`
	TgId     int64  `json:"tg_id" ch:"tg_id"`
	TgUserId int64  `json:"tg_user_id" ch:"tg_user_id"`
	Text     string `json:"text" ch:"text"`
}
