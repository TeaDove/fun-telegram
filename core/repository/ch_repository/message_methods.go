package ch_repository

import (
	"context"
	"time"

	"github.com/pkg/errors"
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

type GroupedCountGetByChatIdByUserIdOutput struct {
	TgUserId int64  `ch:"tg_user_id"`
	Count    uint64 `ch:"count"`
}

func (r *Repository) GroupedCountGetByChatIdByUserId(
	ctx context.Context,
	chatId int64,
	limit int64,
) ([]GroupedCountGetByChatIdByUserIdOutput, error) {
	rows, err := r.conn.Query(ctx, `
select tg_user_id, count(1) as "count"
	from message final
	where tg_chat_id = ?
		group by 1
			order by 2 desc
			limit ?;
`, chatId, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find interlocutors")
	}

	output := make([]GroupedCountGetByChatIdByUserIdOutput, 0, 100)
	for rows.Next() {
		row := GroupedCountGetByChatIdByUserIdOutput{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, err
}

func (r *Repository) CountGetByChatId(
	ctx context.Context,
	chatId int64,
) (uint64, error) {
	row := r.conn.QueryRow(ctx, `
select count() as count from message final
	where tg_chat_id = ?
`, chatId)
	if row.Err() != nil {
		return 0, errors.Wrap(row.Err(), "failed to query row")
	}

	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to scan row")
	}

	return count, nil
}

func (r *Repository) CountGetByChatIdByUserId(
	ctx context.Context,
	chatId int64,
	userId int64,
) (uint64, error) {
	row := r.conn.QueryRow(ctx, `
select count() as count from message final
	where tg_chat_id = ?
		and tg_user_id = ?
`, chatId, userId)
	if row.Err() != nil {
		return 0, errors.Wrap(row.Err(), "failed to query row")
	}

	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "failed to scan row")
	}

	return count, nil
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

func (r *Repository) GetLastMessageByChatId(ctx context.Context, chatId int64) (Message, error) {
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

func (r *Repository) GetLastMessageByChatIdByUserId(ctx context.Context, chatId int64, userId int64) (Message, error) {
	row := r.conn.QueryRow(ctx, `
select id, created_at, tg_chat_id, tg_id, tg_user_id, text from message final
	where tg_chat_id = ? and tg_user_id = ?
		order by created_at limit 1
`, chatId, userId)
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

type MessagesGroupedByTime struct {
	CreatedAt time.Time `ch:"created_at"`
	Count     uint64    `ch:"count"`
}

// GetMessagesGroupedByDateByChatId
// precision = 60 means per minute, 86400 - per day
func (r *Repository) GetMessagesGroupedByDateByChatId(ctx context.Context, chatId int64, precision int) ([]MessagesGroupedByTime, error) {
	rows, err := r.conn.Query(ctx, `
	select fromUnixTimestamp(intDiv(toUnixTimestamp(created_at), ?) * ?) as "created_at", count(1) as "count"
		from message final
			where tg_chat_id = ?
		group by 1
		order by 1 desc;
`, precision, precision, chatId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query rows")
	}

	output := make([]MessagesGroupedByTime, 0, 100)
	for rows.Next() {
		row := MessagesGroupedByTime{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, err
}

func (r *Repository) GetMessagesGroupedByDateByChatIdByUserId(ctx context.Context, chatId int64, userId int64, precision int) ([]MessagesGroupedByTime, error) {
	rows, err := r.conn.Query(ctx, `
	select fromUnixTimestamp(intDiv(toUnixTimestamp(created_at), ?) * ?) as "created_at", count(1) as "count"
		from message final
			where tg_chat_id = ? and tg_user_id = ?
		group by 1
		order by 1 desc;
`, precision, precision, chatId, userId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query rows")
	}

	output := make([]MessagesGroupedByTime, 0, 100)
	for rows.Next() {
		row := MessagesGroupedByTime{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, err
}

// GetMessagesGroupedByTimeByChatId
// precision = 60 means per minute, 86400 - per day
func (r *Repository) GetMessagesGroupedByTimeByChatId(ctx context.Context, chatId int64, precision int) ([]MessagesGroupedByTime, error) {
	rows, err := r.conn.Query(ctx, `
	select toTime(fromUnixTimestamp(intDiv(toUnixTimestamp(created_at), ?) * ?)) as "created_at", count(1) as "count"
		from message final
			where tg_chat_id = ?
		group by 1
		order by 1 desc;
`, precision, precision, chatId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query rows")
	}

	output := make([]MessagesGroupedByTime, 0, 100)
	for rows.Next() {
		row := MessagesGroupedByTime{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, err
}

func (r *Repository) GetMessagesGroupedByTimeByChatIdByUserId(ctx context.Context, chatId int64, userId int64, precision int) ([]MessagesGroupedByTime, error) {
	rows, err := r.conn.Query(ctx, `
	select toTime(fromUnixTimestamp(intDiv(toUnixTimestamp(created_at), ?) * ?)) as "created_at", count(1) as "count"
		from message final
			where tg_chat_id = ? and tg_user_id = ?
		group by 1
		order by 1 desc;
`, precision, precision, chatId, userId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query rows")
	}

	output := make([]MessagesGroupedByTime, 0, 100)
	for rows.Next() {
		row := MessagesGroupedByTime{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, err
}
