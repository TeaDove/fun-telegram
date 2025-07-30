package db_repository

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm/clause"
)

func (r *Repository) ChatUpsert(ctx context.Context, row *Chat) error {
	err := r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "tg_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"tg_id", "title", "updated_at"}),
		}).
		Create(&row).Error
	if err != nil {
		return errors.Wrap(err, "failed to upsert chat")
	}

	return nil
}

func (r *Repository) ChatSelectByID(ctx context.Context, tgID int64) (Chat, error) {
	var chat Chat

	err := r.db.WithContext(ctx).Where("tg_id = ?", tgID).Find(&chat).Limit(1).Error
	if err != nil {
		return Chat{}, errors.Wrap(err, "failed to get chat")
	}

	return chat, nil
}
