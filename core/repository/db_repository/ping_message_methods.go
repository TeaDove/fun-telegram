package db_repository

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

func (r *Repository) PingMessageCreate(ctx context.Context, message *PingMessage) error {
	err := r.db.WithContext(ctx).Create(message).Error
	if err != nil {
		return errors.Wrap(err, "unable to create ping message")
	}

	return nil
}

func (r *Repository) PingMessageGet(ctx context.Context) ([]PingMessage, error) {
	var messages []PingMessage

	err := r.db.WithContext(ctx).Find(&messages).Where("delete_at <= ?", time.Now().UTC()).Error
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping message get")
	}

	return messages, nil
}

func (r *Repository) PingMessageDelete(ctx context.Context, messages []PingMessage) error {
	err := r.db.WithContext(ctx).Delete(messages).Error
	if err != nil {
		return errors.Wrap(err, "failed to ping message delete")
	}

	return nil
}
