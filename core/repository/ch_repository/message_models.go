package ch_repository

import (
	"time"

	"github.com/guregu/null/v5"
)

type Message struct {
	CreatedAt time.Time `csv:"created_at" ch:"created_at"`

	TgChatID        int64  `csv:"tg_chat_id" ch:"tg_chat_id"`
	TgId            int64  `csv:"tg_id" ch:"tg_id"`
	TgUserId        int64  `csv:"tg_user_id" ch:"tg_user_id"`
	Text            string `csv:"text" ch:"text"`
	WordsCount      uint64 `csv:"words_count" ch:"words_count"`
	ToxicWordsCount uint64 `csv:"toxic_words_count" ch:"toxic_words_count"`

	ReplyToMsgID  null.Int64 `csv:"reply_to_msg_id" ch:"reply_to_msg_id"`
	ReplyToUserID null.Int64 `csv:"reply_to_user_id" ch:"reply_to_user_id"`
}
