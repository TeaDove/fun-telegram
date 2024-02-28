package ch_repository

import (
	"time"

	"github.com/guregu/null/v5"

	"github.com/google/uuid"
)

type Message struct {
	Id        uuid.UUID `json:"id" ch:"id"`
	CreatedAt time.Time `json:"created_at" ch:"created_at"`

	TgChatID        int64  `json:"tg_chat_id" ch:"tg_chat_id"`
	TgId            int64  `json:"tg_id" ch:"tg_id"`
	TgUserId        int64  `json:"tg_user_id" ch:"tg_user_id"`
	Text            string `json:"text" ch:"text"`
	WordsCount      uint64 `json:"words_count" ch:"words_count"`
	ToxicWordsCount uint64 `json:"toxic_words_count" ch:"toxic_words_count"`

	ReplyToMsgID  null.Int64 `json:"reply_to_msg_id" ch:"reply_to_msg_id"`
	ReplyToUserID null.Int64 `json:"reply_to_user_id" ch:"reply_to_user_id"`
}
