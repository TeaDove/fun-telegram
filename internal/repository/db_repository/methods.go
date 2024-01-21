package db_repository

import (
	"context"
	errors "github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func (r *Repository) MessageCreateOrNothingAndSetTime(ctx context.Context, message *Message) error {
	message.UpdatedAt = time.Now().UTC()

	_, err := r.messageCollection.InsertOne(ctx, &message)
	if err != nil {
		var mgerr mongo.WriteException
		if errors.As(err, &mgerr) {
			if mgerr.HasErrorCode(11000) {
				return nil
			}
		}

		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) UserUpsert(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now().UTC()

	filter := bson.M{"tg_user_id": user.TgUserId}
	update := bson.M{"$set": bson.M{
		"tg_user_id":  user.TgUserId,
		"tg_username": user.TgUsername,
		"tg_name":     user.TgName,
		"updated_at":  user.UpdatedAt,
	}}
	opts := options.Update().SetUpsert(true)

	result, err := r.userCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return errors.WithStack(err)
	}

	if result.UpsertedCount == 1 {
		_, err = r.userCollection.UpdateOne(ctx,
			bson.M{"tg_user_id": user.TgUserId},
			bson.M{"$set": bson.M{
				"created_at": user.UpdatedAt,
			}})
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (r *Repository) MessageDeleteOld(ctx context.Context) (int64, error) {
	result, err := r.messageCollection.DeleteMany(ctx,
		bson.M{"created_at": bson.M{"$lt": time.Now().UTC().Add(-shared.AppSettings.MessageTtl)}})
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return result.DeletedCount, nil
}

func (r *Repository) GetMessagesByChat(ctx context.Context, chatId int64) ([]Message, error) {
	messages := make([]Message, 0, 100)

	opts := options.Find().SetSort(bson.M{"created_at": -1})

	err := r.messageCollection.SimpleFindWithCtx(ctx, &messages, bson.M{"tg_chat_id": chatId}, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return messages, nil
}

func (r *Repository) GetUsersById(ctx context.Context, usersId []int64) ([]User, error) {
	users := make([]User, 0, len(usersId))

	err := r.userCollection.SimpleFindWithCtx(ctx, &users, bson.M{"tg_user_id": bson.M{"$in": usersId}})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return users, nil
}

func (r *Repository) GetLastMessage(ctx context.Context, chatId int64) (Message, error) {
	var message Message

	err := r.messageCollection.FirstWithCtx(
		ctx,
		bson.M{"tg_chat_id": chatId},
		&message,
		options.FindOne().SetSort(bson.M{"created_at": 1}),
	)
	if err != nil {
		return Message{}, errors.WithStack(err)
	}

	return message, nil
}
