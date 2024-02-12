package ch_repository

import (
	"context"
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
