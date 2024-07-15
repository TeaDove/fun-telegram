package db_repository

import (
	"context"

	"github.com/pkg/errors"
)

func (r *Repository) RestartMessageInsert(
	ctx context.Context,
	restartMessage *RestartMessage,
) error {
	err := r.db.WithContext(ctx).Create(&restartMessage).Error
	if err != nil {
		return errors.Wrap(err, "failed to restart message insert")
	}

	return nil
}

func (r *Repository) RestartMessageGet(ctx context.Context) ([]RestartMessage, error) {
	var messages []RestartMessage

	err := r.db.WithContext(ctx).Find(&messages).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to restart message get")
	}

	return messages, nil
}

func (r *Repository) RestartMessageDelete(ctx context.Context, messages []RestartMessage) error {
	err := r.db.WithContext(ctx).Delete(messages).Error
	if err != nil {
		return errors.Wrap(err, "failed to restart message delete")
	}

	return nil
}
