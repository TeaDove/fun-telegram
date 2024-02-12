package mongo_repository

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (r *Repository) ChatUpsert(ctx context.Context, row *Chat) error {
	row.UpdatedAt = time.Now().UTC()

	filter := bson.M{"tg_id": row.TgId}
	update := bson.M{"$set": bson.M{
		"tg_id":      row.TgId,
		"title":      row.Title,
		"updated_at": row.UpdatedAt,
		"created_at": row.CreatedAt,
	}}
	opts := options.Update().SetUpsert(true)

	_, err := r.chatCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) GetChat(ctx context.Context, chatId int64) (chat Chat, err error) {
	err = r.chatCollection.FirstWithCtx(ctx, bson.M{"tg_id": chatId}, &chat)
	if err != nil {
		return Chat{}, errors.Wrap(err, "failed to get chat")
	}

	return chat, nil
}
