package analitics

import (
	"time"

	"github.com/guregu/null/v5"
)

type Message struct {
	CreatedAt time.Time `json:"created_at"`

	TgChatID      int64      `json:"tg_chat_id"`
	TgID          int        `json:"tg_id"`
	TgUserID      int64      `json:"tg_user_id"`
	Text          string     `json:"text"`
	ReplyToMsgID  null.Int64 `json:"reply_to_msg_id"`
	ReplyToUserID null.Int64 `json:"reply_to_user_id"`
}
