package db_repository

import (
	"time"

	"github.com/guregu/null/v5"
)

type Message struct {
	WithId
	WithCreatedAt

	TgChatID int64 `sql:"tg_chat_id" gorm:"index:tg_chat_id_tg_id_idx,unique"`
	TgId     int   `sql:"tg_id"      gorm:"index:tg_chat_id_tg_id_idx,unique"`

	TgUserId        int64  `sql:"tg_user_id"        gorm:"index"`
	Text            string `sql:"text"`
	WordsCount      uint64 `sql:"words_count"`
	ToxicWordsCount uint64 `sql:"toxic_words_count"`

	ReplyToTgMsgID  null.Int64 `sql:"reply_to_tg_msg_id"  gorm:"index"`
	ReplyToTgUserID null.Int64 `sql:"reply_to_tg_user_id" gorm:"index"`
}

type MessageGroupByChatIdAndUserIdOutput struct {
	TgUserId        int64  `sql:"tg_user_id"`
	WordsCount      uint64 `sql:"words_count"`
	ToxicWordsCount uint64 `sql:"toxic_words_count"`
}

type MessageGroupByTimeOutput struct {
	CreatedAt  time.Time `sql:"created_at"`
	WordsCount uint64    `sql:"words_count"`
}

type MessagesGroupByTimeByWeekdayOutput struct {
	IsWeekend  bool      `sql:"is_weekend"`
	CreatedAt  time.Time `sql:"created_at"`
	WordsCount uint64    `sql:"words_count"`
}

type MessageGroupByInterlocutorsOutput struct {
	TgUserId      int64  `sql:"tg_user_id"`
	MessagesCount uint64 `sql:"count"`
}
