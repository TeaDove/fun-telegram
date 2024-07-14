package db_repository

import (
	"context"
	"github.com/pkg/errors"
)

func (r *Repository) MessageInsert(ctx context.Context, input *Message) error {
	err := r.db.Create(input).WithContext(ctx).Error
	if err != nil {
		return errors.Wrap(err, "failed to insert message")
	}

	return nil
}

func (r *Repository) MessagesDeleteByChat(ctx context.Context, tgChatId int64) (uint64, error) {
	resp := r.db.Delete(&Message{}, "tg_chat_id = ?", tgChatId).WithContext(ctx)
	if resp.Error != nil {
		return 0, errors.Wrap(resp.Error, "failed to delete message")
	}

	return uint64(resp.RowsAffected), nil
}

func (r *Repository) MessageCountByChatId(
	ctx context.Context,
	tgChatId int64,
) (uint64, error) {
	var count int64
	err := r.db.
		WithContext(ctx).
		Model(&Message{}).
		Where("tg_chat_id = ?", tgChatId).
		Count(&count).
		Error

	if err != nil {
		return 0, errors.Wrap(err, "failed to count message")
	}

	return uint64(count), nil
}

func (r *Repository) MessageCountByChatIdAndUserId(
	ctx context.Context,
	tgChatId int64,
	tgUserId int64,
) (uint64, error) {
	var count int64
	err := r.db.
		WithContext(ctx).
		Model(&Message{}).
		Where("tg_chat_id = ? AND tg_user_id = ?", tgChatId, tgUserId).
		Count(&count).
		Error

	if err != nil {
		return 0, errors.Wrap(err, "failed to count message")
	}

	return uint64(count), nil
}

func (r *Repository) MessageGroupByChatIdAndUserId(
	ctx context.Context,
	tgChatId int64,
	tgUserIds []int64,
	limit int64,
	desc bool,
) ([]MessageGroupByChatIdAndUserIdOutput, error) {
	var output []MessageGroupByChatIdAndUserIdOutput

	const queryByAsc = `
select 
    tg_user_id, 
    sum(words_count) as "words_count",
	sum(toxic_words_count) as "toxic_words_count"
from messages where tg_chat_id = ? and tg_user_id in (?)
group by 1
order by 2 asc
limit ?
`
	const queryByDesc = `
select 
    tg_user_id, 
    sum(words_count) as "words_count",
	sum(toxic_words_count) as "toxic_words_count"
from messages where tg_chat_id = ? and tg_user_id in (?)
group by 1
order by 2 desc 
limit ?
`
	var query string
	if desc {
		query = queryByDesc
	} else {
		query = queryByAsc
	}

	err := r.db.WithContext(ctx).Raw(query, tgChatId, tgUserIds, limit).Scan(&output).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to group by messages")
	}

	return output, nil
}

func (r *Repository) MessageGetLastByChatId(ctx context.Context, tgChatId int64) (Message, error) {
	var message Message

	err := r.db.
		WithContext(ctx).
		Model(message).
		Where("tg_chat_id = ?", tgChatId).
		Order("created_at desc").
		Limit(1).
		Find(&message).Error
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to get message")
	}

	return message, nil
}

func (r *Repository) MessageGetLastByChatIdAndUserId(ctx context.Context, tgChatId int64, tgUserId int64) (Message, error) {
	var message Message

	err := r.db.
		WithContext(ctx).
		Model(message).
		Where("tg_chat_id = ? and tg_user_id = ?", tgChatId, tgUserId).
		Order("created_at desc").
		Limit(1).
		Find(&message).Error
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to get message")
	}

	return message, nil
}
