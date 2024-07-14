package db_repository

import (
	"github.com/guregu/null/v5"
)

type Message struct {
	WithId
	WithCreatedAt

	TgChatID        int64  `sql:"tg_chat_id"`
	TgId            int64  `sql:"tg_id"`
	TgUserId        int64  `sql:"tg_user_id"`
	Text            string `sql:"text"`
	WordsCount      uint64 `sql:"words_count"`
	ToxicWordsCount uint64 `sql:"toxic_words_count"`

	ReplyToMsgID  null.Int64 `sql:"reply_to_msg_id"`
	ReplyToUserID null.Int64 `sql:"reply_to_user_id"`
}

type RestartMessage struct {
	WithId
	WithCreatedAt

	MessageId uint `sql:"message_id"`
}

type PingMessage struct {
	WithId
	WithCreatedAt

	MessageId uint `sql:"message_id"`
}

type MessageGroupByChatIdAndUserIdOutput struct {
	TgUserId        int64  `ch:"tg_user_id"`
	WordsCount      uint64 `ch:"words_count"`
	ToxicWordsCount uint64 `ch:"toxic_words_count"`
}
