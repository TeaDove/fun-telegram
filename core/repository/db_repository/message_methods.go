package db_repository

import (
	"context"
	"time"

	"gorm.io/gorm/clause"

	"gorm.io/gorm"

	"github.com/pkg/errors"
)

func (r *Repository) messageSetReplyToUserId(
	ctx context.Context,
	tx *gorm.DB,
	input *Message,
) error {
	if input.ReplyToTgMsgID.Valid {
		err := tx.WithContext(ctx).
			Model(&Message{}).
			Where("reply_to_tg_msg_id = ? AND tg_chat_id = ?", input.ReplyToTgMsgID.Int64, input.TgChatID).
			Update("reply_to_tg_user_id", input.TgUserId).
			Error
		if err != nil {
			return errors.Wrap(err, "failed to update reply_to_user_id")
		}
	}

	// TODO set replied message
	return nil
}

func (r *Repository) MessageInsert(ctx context.Context, input *Message) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.WithContext(ctx).Clauses(
			clause.OnConflict{
				Columns: []clause.Column{{Name: "tg_chat_id"}, {Name: "tg_id"}},
				DoUpdates: clause.AssignmentColumns(
					[]string{"text", "words_count", "toxic_words_count"},
				),
			}).Create(input).Error
		if err != nil {
			return errors.Wrap(err, "failed to insert message")
		}

		err = r.messageSetReplyToUserId(ctx, tx, input)
		if err != nil {
			return errors.Wrap(err, "failed to set reply to user id")
		}

		return nil
	})
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
from message where tg_chat_id = ? and tg_user_id in (?)
group by 1
order by 2 asc
limit ?
`
	const queryByDesc = `
select 
    tg_user_id, 
    sum(words_count) as "words_count",
	sum(toxic_words_count) as "toxic_words_count"
from message where tg_chat_id = ? and tg_user_id in (?)
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
		Find(&message).
		Where("tg_chat_id = ?", tgChatId).
		Order("created_at desc").
		Limit(1).
		Error
	if err != nil {
		return Message{}, errors.Wrap(err, "failed to get message")
	}

	return message, nil
}

func (r *Repository) MessageGroupByDateAndChatId(
	ctx context.Context,
	tgChatId int64,
	precision time.Duration,
) ([]MessageGroupByTimeOutput, error) {
	var output []MessageGroupByTimeOutput

	precisionSeconds := int(precision.Seconds())

	err := r.db.WithContext(ctx).Raw(`
select 
    datetime((unixepoch(m.created_at) / ?) * ?, 'unixepoch') as "created_at", 
    sum(m.words_count) as "words_count"
from message m 
where tg_chat_id = ?
group by 1
order by 1 desc
`, precisionSeconds, precisionSeconds, tgChatId).
		Scan(&output).
		Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to group by messages")
	}

	return output, nil
}

func (r *Repository) MessageFindRepliesTo(
	ctx context.Context,
	tgChatId int64,
	tgUserId int64,
	minReplyCount int,
	limit int,
) ([]MessageGroupByInterlocutorsOutput, error) {
	var output []MessageGroupByInterlocutorsOutput

	err := r.db.WithContext(ctx).Raw(`
select 
    am.reply_to_tg_user_id as tg_user_id, 
    count(1) as count
	from message am
where am.tg_chat_id = ? and am.tg_user_id = ? and am.reply_to_tg_user_id is not null
group by 1
	having count(1) > ?
order by 2 desc limit ?
`, tgChatId, tgUserId, minReplyCount, limit).Scan(&output).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to group by messages")
	}

	return output, nil
}
