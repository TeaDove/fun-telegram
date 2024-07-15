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

func (r *Message) ToParquet() MessageParquet {
	message := MessageParquet{
		CreatedAt:       r.CreatedAt.UnixNano() / int64(time.Millisecond),
		TgChatID:        r.TgChatID,
		TgId:            int64(r.TgId),
		TgUserId:        r.TgUserId,
		Text:            r.Text,
		WordsCount:      int64(r.WordsCount),
		ToxicWordsCount: int64(r.ToxicWordsCount),
	}

	if r.ReplyToTgMsgID.Valid {
		message.ReplyToTgMsgID = &r.ReplyToTgMsgID.Int64
	}

	if r.ReplyToTgUserID.Valid {
		message.ReplyToTgUserID = &r.ReplyToTgUserID.Int64
	}

	return message
}

type MessageParquet struct {
	CreatedAt int64 `parquet:"name=created_at, type=INT64, convertedtype=TIMESTAMP_MILLIS"`

	TgChatID        int64  `parquet:"name=tg_chat_id, type=INT64"`
	TgId            int64  `parquet:"name=tg_id, type=INT64"`
	TgUserId        int64  `parquet:"name=tg_user_id, type=INT64"`
	Text            string `parquet:"name=text, type=BYTE_ARRAY, convertedtype=UTF8, encoding=PLAIN_DICTIONARY"`
	WordsCount      int64  `parquet:"name=words_count, type=INT64"`
	ToxicWordsCount int64  `parquet:"name=toxic_words_count, type=INT64"`

	ReplyToTgMsgID  *int64 `parquet:"name=reply_to_tg_msg_id, type=INT64"`
	ReplyToTgUserID *int64 `parquet:"name=reply_to_tg_user_id, type=INT64"`
}

type RestartMessage struct {
	WithId
	WithCreatedAt

	MessageTgChatID int64 `sql:"message_tg_chat_id" gorm:"index:tg_chat_id_tg_id_idx,unique"`
	MessageTgId     int   `sql:"message_tg_id"      gorm:"index:tg_chat_id_tg_id_idx,unique"`
}

type PingMessage struct {
	WithId
	WithCreatedAt

	MessageTgChatID int64     `sql:"message_tg_chat_id" gorm:"index:tg_chat_id_tg_id_idx,unique"`
	MessageTgId     int       `sql:"message_tg_id"      gorm:"index:tg_chat_id_tg_id_idx,unique"`
	DeleteAt        time.Time `sql:"delete_at"          gorm:"index"`
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
