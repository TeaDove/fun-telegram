package ch_repository

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

func (r *Repository) MessageCreate(ctx context.Context, message *Message) error {
	err := r.conn.AsyncInsert(ctx, `INSERT INTO message VALUES (
			?, ?, ?, ?, ?, ?
		)`, false, message.Id, message.CreatedAt, message.TgChatID, message.TgId, message.TgUserId, message.Text)
	if err != nil {
		return errors.Wrap(err, "failed to async insert")
	}

	return nil
}

func (r *Repository) MessageDeleteByChatId(ctx context.Context, chatId int64) error {
	err := r.conn.Exec(ctx, `DELETE FROM message WHERE tg_chat_id = ?`, chatId)
	if err != nil {
		return errors.Wrap(err, "failed to delete messages by chat")
	}

	return err
}

type MessageFindInterlocutorsOutput struct {
	TgUserId      int64  `ch:"tg_user_id"`
	MessagesCount uint64 `ch:"count"`
}

func (r *Repository) MessageFindInterlocutors(
	ctx context.Context,
	chatId int64,
	userId int64,
	limit int,
	interlocutorLimit time.Duration,
) ([]MessageFindInterlocutorsOutput, error) {
	rows, err := r.conn.Query(ctx, `
select m.tg_user_id as tg_user_id, count(1) as count
from message am final
         join default.message m
              on am.tg_chat_id = m.tg_chat_id
	where am.tg_chat_id = ?
	  and am.tg_user_id = ?
	  and abs(am.created_at - m.created_at) - ? < 0
	  and am.tg_user_id != m.tg_user_id
	group by 1
	order by 2 desc 
	limit ?
`, chatId, userId, int(interlocutorLimit.Seconds()), limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find interlocutors")
	}

	output := make([]MessageFindInterlocutorsOutput, 0, limit)
	for rows.Next() {
		row := MessageFindInterlocutorsOutput{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, err
}

func (r *Repository) MessageGetByChatIdAndUserId(
	ctx context.Context,
	chatId int64,
	userId int64,
) ([]Message, error) {
	rows, err := r.conn.Query(ctx, `
select id, created_at, tg_chat_id, tg_id, tg_user_id, text from message final
	where tg_chat_id = ?
		and tg_user_id = ?
`, chatId, userId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find interlocutors")
	}

	output := make([]Message, 0, 100)
	for rows.Next() {
		row := Message{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, err
}

func (r *Repository) MessageGetByChatId(
	ctx context.Context,
	chatId int64,
) ([]Message, error) {
	rows, err := r.conn.Query(ctx, `
select id, created_at, tg_chat_id, tg_id, tg_user_id, text from message final
	where tg_chat_id = ?
`, chatId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find interlocutors")
	}

	output := make([]Message, 0, 100)
	for rows.Next() {
		row := Message{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, err
}

func (r *Repository) GetLastMessage(ctx context.Context, chatId int64) (Message, error) {
	row := r.conn.QueryRow(ctx, `
select id, created_at, tg_chat_id, tg_id, tg_user_id, text from message final
	where tg_chat_id = ? 
		order by created_at limit 1
`, chatId)
	if row.Err() != nil {
		return Message{}, errors.Wrap(row.Err(), "failed to select row from clickhouse")
	}

	var message Message
	err := row.ScanStruct(&message)
	if err != nil {
		return Message{}, errors.Wrap(row.Err(), "failed to scan row")
	}

	return message, err
}
