package ch_repository

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

func (r *Repository) MessageInsert(ctx context.Context, message *Message) error {
	err := r.conn.AsyncInsert(ctx, `
INSERT INTO message VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?
		)`,
		false,
		message.CreatedAt,
		message.TgChatID,
		message.TgId,
		message.TgUserId,
		message.Text,
		message.ReplyToMsgID,
		message.ReplyToUserID,
		message.WordsCount,
		message.ToxicWordsCount,
	)
	if err != nil {
		return errors.Wrap(err, "failed to async insert")
	}

	return nil
}

func (r *Repository) MessageSetReplyToUserId(ctx context.Context, chatId int64) error {
	err := r.conn.AsyncInsert(ctx, `
insert into message
select am.created_at,
       am.tg_chat_id,
       am.tg_id,
       am.tg_user_id,
       am.text,
       am.reply_to_msg_id,
       m.tg_user_id,
       am.words_count,
       am.toxic_words_count
from message am
         join default.message m on m.tg_chat_id = am.tg_chat_id AND m.tg_id = am.reply_to_msg_id
where reply_to_msg_id is not null and reply_to_user_id is null and am.tg_chat_id = ?;
`, false, chatId)
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

	return nil
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

	return output, nil
}

func (r *Repository) MessageFindRepliesTo(
	ctx context.Context,
	chatId int64,
	userId int64,
	minReplyCount int,
	limit int,
) ([]MessageFindInterlocutorsOutput, error) {
	rows, err := r.conn.Query(ctx, `
select am.reply_to_user_id as tg_user_id, count(1) as count
	from message am final
where am.tg_chat_id = ? and am.tg_user_id = ? and am.reply_to_user_id != 0 and am.reply_to_user_id is not null
group by 1
	having count(1) > ?
order by 2 desc limit ?
`, chatId, userId, minReplyCount, limit)
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

	return output, nil
}

func (r *Repository) MessageFindRepliedBy(
	ctx context.Context,
	chatId int64,
	userId int64,
	minReplyCount int,
	limit int,
) ([]MessageFindInterlocutorsOutput, error) {
	rows, err := r.conn.Query(ctx, `
select am.tg_user_id as tg_user_id, count(1) as count
	from message am final
		where am.tg_chat_id = ? and am.reply_to_user_id = ?
	group by 1
		having count(1) > ?
		order by 2 desc
		LIMIT ?;
`, chatId, userId, minReplyCount, limit)
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

	return output, nil
}

type MessageFindAllRepliedByOutput struct {
	TgUserId        int64  `ch:"tg_user_id"`
	RepliedTgUserId int64  `ch:"replied_tg_user_id"`
	Count           uint64 `ch:"count"`
}

func (r *Repository) MessageFindAllRepliedBy(
	ctx context.Context,
	chatId int64,
	limit int,
) ([]MessageFindAllRepliedByOutput, error) {
	rows, err := r.conn.Query(ctx, `
select am.tg_user_id as tg_user_id, am.reply_to_user_id as replied_tg_user_id, count(1) as count
	from message am final
		where am.reply_to_user_id != am.tg_user_id and tg_chat_id = ? and m.reply_to_user_id != 0 and m.reply_to_user_id is not null
			group by 1, 2
			order by 3 desc
	limit ?;
`, chatId, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find interlocutors")
	}

	output := make([]MessageFindAllRepliedByOutput, 0, limit)
	for rows.Next() {
		row := MessageFindAllRepliedByOutput{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, nil
}

func (r *Repository) MessageGetByChatIdAndUserId(
	ctx context.Context,
	chatId int64,
	userId int64,
) ([]Message, error) {
	rows, err := r.conn.Query(ctx, `
select created_at, tg_chat_id, tg_id, tg_user_id, text, reply_to_msg_id, reply_to_user_id from message final
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

	return output, nil
}

type GroupedCountGetByChatIdByUserIdOutput struct {
	TgUserId        int64  `ch:"tg_user_id"`
	WordsCount      uint64 `ch:"words_count"`
	ToxicWordsCount uint64 `ch:"toxic_words_count"`
}

func (r *Repository) GroupedCountGetByChatIdByUserId(
	ctx context.Context,
	chatId int64,
	limit int64,
) ([]GroupedCountGetByChatIdByUserIdOutput, error) {
	rows, err := r.conn.Query(ctx, `
select tg_user_id, sum(words_count) as "words_count", sum(toxic_words_count) as "toxic_words_count"
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

	return output, nil
}

func (r *Repository) GroupedCountGetByChatIdByUserIdAsc(
	ctx context.Context,
	chatId int64,
	limit int64,
	userIds []int64,
) ([]GroupedCountGetByChatIdByUserIdOutput, error) {
	rows, err := r.conn.Query(ctx, `
with "user" as
             (select arrayJoin(cast(?, 'Array(Int64)')) as "id")
select u.id as "tg_user_id", sum(words_count) as "words_count", sum(toxic_words_count) as "toxic_words_count"
	from "user" u
         left join message m on m.tg_user_id = u.id
where tg_chat_id = ?
	group by 1
		order by 2 asc
		limit ?;
`, userIds, chatId, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select messages with cte")
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

	return output, nil
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
select created_at, tg_chat_id, tg_id, tg_user_id, text, reply_to_msg_id, reply_to_user_id from message final
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

	return output, nil
}

func (r *Repository) GetLastMessageByChatId(ctx context.Context, chatId int64) (Message, error) {
	row := r.conn.QueryRow(ctx, `
select created_at, tg_chat_id, tg_id, tg_user_id, text, reply_to_msg_id, reply_to_user_id from message final
	where tg_chat_id = ? 
		order by created_at limit 1
`, chatId)
	if row.Err() != nil {
		return Message{}, errors.Wrap(row.Err(), "failed to select row from clickhouse")
	}

	var message Message
	err := row.ScanStruct(&message)
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to scan row")
	}

	return message, nil
}

func (r *Repository) GetLastMessageByChatIdByUserId(ctx context.Context, chatId int64, userId int64) (Message, error) {
	row := r.conn.QueryRow(ctx, `
select created_at, tg_chat_id, tg_id, tg_user_id, text, reply_to_msg_id, reply_to_user_id from message final
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
	CreatedAt  time.Time `ch:"created_at"`
	WordsCount uint64    `ch:"words_count"`
}

type MessagesGroupedByTimeByWeekday struct {
	IsWeekend  bool      `ch:"is_weekend"`
	CreatedAt  time.Time `ch:"created_at"`
	WordsCount uint64    `ch:"words_count"`
}

// GetMessagesGroupedByDateByChatId
// precision = 60 means per minute, 86400 - per day
func (r *Repository) GetMessagesGroupedByDateByChatId(ctx context.Context, chatId int64, precision int) ([]MessagesGroupedByTime, error) {
	rows, err := r.conn.Query(ctx, `
	select fromUnixTimestamp(intDiv(toUnixTimestamp(created_at), ?) * ?) as "created_at", 
	       sum(words_count) as "words_count"
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

	return output, nil
}

func (r *Repository) GetMessagesGroupedByDateByChatIdByUserId(ctx context.Context, chatId int64, userId int64, precision int) ([]MessagesGroupedByTime, error) {
	rows, err := r.conn.Query(ctx, `
	select fromUnixTimestamp(intDiv(toUnixTimestamp(created_at), ?) * ?) as "created_at", sum(m.words_count) as "words_count"
		from message m final
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

	return output, nil
}

// GetMessagesGroupedByTimeByChatId
// precision = 60 means per minute, 86400 - per day
func (r *Repository) GetMessagesGroupedByTimeByChatId(
	ctx context.Context,
	chatId int64,
	precision int,
	tz int8,
) ([]MessagesGroupedByTimeByWeekday, error) {
	rows, err := r.conn.Query(ctx, `
	select case when toDayOfWeek(m.created_at + interval ? hour) >= 6 then true else false end as is_weekend,
	       toTime(fromUnixTimestamp(intDiv(toUnixTimestamp(created_at + interval ? hour), ?) * ?)) as created_at, 
	       sum(m.words_count) as words_count
		from message m final
			where tg_chat_id = ?
		group by 1, 2
		order by 1 desc;
`, tz, tz, precision, precision, chatId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query rows")
	}

	output := make([]MessagesGroupedByTimeByWeekday, 0, 100)
	for rows.Next() {
		row := MessagesGroupedByTimeByWeekday{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, nil
}

func (r *Repository) GetMessagesGroupedByTimeByChatIdByUserId(
	ctx context.Context,
	chatId int64,
	userId int64,
	precision int,
	tz int8,
) ([]MessagesGroupedByTimeByWeekday, error) {
	rows, err := r.conn.Query(ctx, `
	select case when toDayOfWeek(m.created_at + interval ? hour) >= 6 then true else false end as is_weekend,
	       toTime(fromUnixTimestamp(intDiv(toUnixTimestamp(created_at + interval ? hour), ?) * ?)) as "created_at", 
	       sum(m.words_count) as "words_count"
		from message m final
			where tg_chat_id = ? and tg_user_id = ? 
		group by 1, 2
		order by 1 desc;
`, tz, tz, precision, precision, chatId, userId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query rows")
	}

	output := make([]MessagesGroupedByTimeByWeekday, 0, 100)
	for rows.Next() {
		row := MessagesGroupedByTimeByWeekday{}
		err = rows.ScanStruct(&row)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		output = append(output, row)
	}

	return output, nil
}

func (r *Repository) MessagesGetByChatIds(ctx context.Context, tgChatIds []int64) ([]Message, error) {
	rows, err := r.conn.Query(ctx, `
select m.created_at,
       m.tg_chat_id,
       m.tg_id,
       m.tg_user_id,
       m.text,
       m.reply_to_msg_id,
       m.reply_to_user_id,
       m.words_count,
       m.toxic_words_count
from message m final
    where m.tg_chat_id in ?
`, tgChatIds)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select messages")
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

	return output, nil
}
