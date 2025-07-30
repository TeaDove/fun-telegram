package db_repository

import (
	"github.com/guregu/null/v5"
)

type Message struct {
	WithID
	WithCreatedAt

	TgChatID int64 `sql:"tg_chat_id" gorm:"index:tg_chat_id_tg_id_idx,unique"`
	TgID     int   `sql:"tg_id"      gorm:"index:tg_chat_id_tg_id_idx,unique"`

	TgUserID        int64  `sql:"tg_user_id"        gorm:"index"`
	Text            string `sql:"text"`
	WordsCount      uint64 `sql:"words_count"`
	ToxicWordsCount uint64 `sql:"toxic_words_count"`

	ReplyToTgMsgID  null.Int64 `sql:"reply_to_tg_msg_id"  gorm:"index"`
	ReplyToTgUserID null.Int64 `sql:"reply_to_tg_user_id" gorm:"index"`
}

type MessageGroupByChatIDAndUserIDOutput struct {
	TgUserID        int64  `sql:"tg_user_id"`
	WordsCount      uint64 `sql:"words_count"`
	ToxicWordsCount uint64 `sql:"toxic_words_count"`
}

type MessageGroupByTimeOutput struct {
	CreatedAt  string `sql:"created_at"`
	WordsCount uint64 `sql:"words_count"`
}

type MessagesGroupByTimeByWeekdayOutput struct {
	IsWeekend  bool   `sql:"is_weekend"`
	CreatedAt  string `sql:"created_at"`
	WordsCount uint64 `sql:"words_count"`
}

type MessageGroupByInterlocutorsOutput struct {
	TgUserID      int64  `sql:"tg_user_id"`
	MessagesCount uint64 `sql:"count"`
}
