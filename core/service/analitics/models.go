package analitics

import "time"

type Message struct {
	CreatedAt time.Time `json:"created_at"`

	TgChatID int64  `json:"tg_chat_id"`
	TgId     int64  `json:"tg_id"`
	TgUserId int64  `json:"tg_user_id"`
	Text     string `json:"text"`
}
