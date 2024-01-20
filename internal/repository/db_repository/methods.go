package db_repository

import (
	"context"
	errors "github.com/pkg/errors"
	"github.com/rs/zerolog"
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

func (r *Repository) MessageDeleteOld(ctx context.Context) error {
	const messageTtl = 6 * 30 * 24 * time.Hour // 6 months

	result, err := r.messageCollection.DeleteMany(ctx,
		bson.D{{"created_at",
			bson.D{{"$lt", time.Now().UTC().Add(-messageTtl)}},
		}})
	if err != nil {
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx).Info().Str("status", "old.messages.deleted").Int64("count", result.DeletedCount).Send()
	return nil
}

func (r *Repository) GetMessagesByChat(ctx context.Context, chatId int64) ([]Message, error) {
	messages := make([]Message, 0, 100)

	opts := options.Find().SetSort(bson.D{{"created_at", -1}})
	err := r.messageCollection.SimpleFindWithCtx(ctx, &messages, bson.D{{"tg_chat_id", chatId}}, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return messages, nil
}

func (r *Repository) GetUsersById(ctx context.Context, usersId []int64) ([]User, error) {
	users := make([]User, 0, len(usersId))

	err := r.userCollection.SimpleFindWithCtx(ctx, &users, bson.D{{"tg_user_id", bson.D{{"$in", usersId}}}})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return users, nil
}
