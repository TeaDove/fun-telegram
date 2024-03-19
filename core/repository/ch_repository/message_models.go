package ch_repository

import (
	"time"

	"github.com/guregu/null/v5"
)

type Message struct {
	CreatedAt time.Time `csv:"created_at" ch:"created_at"`

	TgChatID        int64  `csv:"tg_chat_id"        ch:"tg_chat_id"`
	TgId            int64  `csv:"tg_id"             ch:"tg_id"`
	TgUserId        int64  `csv:"tg_user_id"        ch:"tg_user_id"`
	Text            string `csv:"text"              ch:"text"`
	WordsCount      uint64 `csv:"words_count"       ch:"words_count"`
	ToxicWordsCount uint64 `csv:"toxic_words_count" ch:"toxic_words_count"`

	ReplyToMsgID  null.Int64 `csv:"reply_to_msg_id"  ch:"reply_to_msg_id"`
	ReplyToUserID null.Int64 `csv:"reply_to_user_id" ch:"reply_to_user_id"`
}

func (r *Message) ToParquet() MessageParquet {
	message := MessageParquet{
		CreatedAt:       r.CreatedAt.UnixNano() / int64(time.Millisecond),
		TgChatID:        r.TgChatID,
		TgId:            r.TgId,
		TgUserId:        r.TgUserId,
		Text:            r.Text,
		WordsCount:      int64(r.WordsCount),
		ToxicWordsCount: int64(r.ToxicWordsCount),
	}

	if r.ReplyToMsgID.Valid {
		message.ReplyToMsgID = &r.ReplyToMsgID.Int64
	}

	if r.ReplyToUserID.Valid {
		message.ReplyToUserID = &r.ReplyToUserID.Int64
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

	ReplyToMsgID  *int64 `parquet:"name=reply_to_msg_id, type=INT64"`
	ReplyToUserID *int64 `parquet:"name=reply_to_user_id, type=INT64"`
}
