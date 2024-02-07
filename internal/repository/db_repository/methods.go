package db_repository

import (
	"context"
	"github.com/kamva/mgm/v3/builder"
	"github.com/kamva/mgm/v3/operator"
	errors "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (r *Repository) MessageGetSortedLimited(ctx context.Context, limit int64) ([]Message, error) {
	messages := make([]Message, 0, 100)

	opts := options.Find().SetSort(bson.M{"created_at": 1}).SetLimit(limit)

	err := r.messageCollection.SimpleFindWithCtx(ctx, &messages, bson.M{}, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return messages, nil
}

func (r *Repository) DeleteMessages(ctx context.Context, messages []Message) (int64, error) {
	messageIds := make([]primitive.ObjectID, len(messages))
	for idx, message := range messages {
		messageIds[idx] = message.ID
	}

	result, err := r.messageCollection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": messageIds}})
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return result.DeletedCount, nil
}

func (r *Repository) DeleteMessagesOldWithCount(ctx context.Context, limit int64) (int64, error) {
	batchSize := int64(10_000)
	count := int64(0)
	for {
		shouldBreak := false
		if batchSize+count > limit {
			batchSize = limit - count
			shouldBreak = true
		}

		messages, err := r.MessageGetSortedLimited(ctx, batchSize)
		if err != nil {
			return 0, errors.WithStack(err)
		}

		batchCount, err := r.DeleteMessages(ctx, messages)
		if err != nil {
			return 0, errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "messages.deleted").Int64("count", batchCount).Send()

		count += batchCount

		if shouldBreak {
			break
		}
	}

	return count, nil
}

func (r *Repository) MessageDeleteOld(ctx context.Context) (int64, error) {
	result, err := r.messageCollection.DeleteMany(ctx,
		bson.M{"created_at": bson.M{"$lt": time.Now().UTC().Add(-time.Hour * 24 * 365)}})
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

func (r *Repository) GetMessagesByChatAndUsername(
	ctx context.Context,
	chatId int64,
	username string,
) ([]Message, error) {
	messages := make([]Message, 0, 100)

	err := r.messageCollection.SimpleAggregateWithCtx(
		ctx,
		&messages,
		builder.Lookup(r.userCollection.Name(), "tg_user_id", "tg_user_id", "user"),
		bson.M{
			operator.Project: bson.M{
				"username":   "$user.tg_username",
				"text":       1,
				"tg_chat_id": 1,
				"tg_id":      1,
				"created_at": 1,
				"updated_at": 1,
			},
		},
		bson.M{operator.Unwind: "$username"},
		bson.M{operator.Match: bson.M{"username": username, "tg_chat_id": chatId}},
	)
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

func (r *Repository) GetUsersByChatId(ctx context.Context, chatId int64) ([]User, error) {
	userIds, err := r.messageCollection.Distinct(ctx, "tg_user_id", bson.M{"tg_chat_id": chatId})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	usersIdsConcrete := make([]int64, 0, len(userIds))

	for _, userId := range userIds {
		userIdConcrete, ok := userId.(int64)
		if !ok {
			return nil, errors.New("non int64 type")
		}

		usersIdsConcrete = append(usersIdsConcrete, userIdConcrete)
	}

	return r.GetUsersById(ctx, usersIdsConcrete)
}

func (r *Repository) GetUserById(ctx context.Context, userId int64) (User, error) {
	var user User

	err := r.userCollection.FirstWithCtx(ctx, bson.M{"tg_user_id": userId}, &user)
	if err != nil {
		return User{}, errors.WithStack(err)
	}

	return user, nil
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

func (r *Repository) CheckUserExists(ctx context.Context, userId int64) (bool, error) {
	count, err := r.userCollection.CountDocuments(ctx, bson.M{"tg_user_id": userId})
	if err != nil {
		return false, errors.WithStack(err)
	}

	return count == 1, nil
}

func (r *Repository) DeleteMessagesByChat(ctx context.Context, chatId int64) (int64, error) {
	result, err := r.messageCollection.DeleteMany(ctx, bson.M{"tg_chat_id": chatId})
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return result.DeletedCount, nil
}

func (r *Repository) DeleteAllMessages(ctx context.Context) (int64, error) {
	result, err := r.messageCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return result.DeletedCount, nil
}
