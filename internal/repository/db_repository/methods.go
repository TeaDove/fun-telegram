package db_repository

import (
	"context"
	errors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (r *Repository) MessageCreate(ctx context.Context, message *Message) error {
	err := r.messageCollection.CreateWithCtx(ctx, message)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) UserUpsert(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now().UTC()

	filter := bson.D{{"tg_user_id", user.TgUserId}}
	update := bson.D{{"$set", bson.D{
		{"tg_user_id", user.TgUserId},
		{"tg_username", user.TgUsername},
		{"tg_name", user.TgName},
		{"updated_at", user.UpdatedAt},
	}}}
	opts := options.Update().SetUpsert(true)

	result, err := r.userCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return errors.WithStack(err)
	}

	if result.UpsertedCount == 1 {
		_, err = r.userCollection.UpdateOne(ctx,
			bson.D{{"tg_user_id", user.TgUserId}},
			bson.D{{"$set", bson.D{
				{"created_at", user.UpdatedAt},
			}}})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
