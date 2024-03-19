package mongo_repository

import (
	"context"
	"time"

	"github.com/kamva/mgm/v3/builder"
	"github.com/kamva/mgm/v3/operator"
	errors "github.com/pkg/errors"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const duplicationError = 11000

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
			if mgerr.HasErrorCode(duplicationError) {
				return nil
			}
		}

		return errors.WithStack(err)
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

// GetMessagesByChat
// Deprecated
func (r *Repository) GetMessagesByChat(ctx context.Context, chatId int64) ([]Message, error) {
	messages := make([]Message, 0, 100)

	opts := options.Find().SetSort(bson.M{"created_at": -1})

	err := r.messageCollection.SimpleFindWithCtx(ctx, &messages, bson.M{"tg_chat_id": chatId}, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return messages, nil
}

// GetMessagesByChatAndUsername
// Deprecated
func (r *Repository) GetMessagesByChatAndUsername(
	ctx context.Context,
	chatId int64,
	username string,
) ([]Message, error) {
	messages := make([]Message, 0, 100)

	err := r.messageCollection.SimpleAggregateWithCtx(
		ctx,
		&messages,
		builder.Lookup(r.userCollection.Name(), "tg_user_id", "tg_id", "user"),
		bson.M{operator.Unwind: "$user"},
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
		bson.M{operator.Match: bson.M{"username": username, "tg_chat_id": chatId}},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return messages, nil
}

// GetLastMessage
// Deprecated
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

// CheckUserExists
// Deprecated
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

func (r *Repository) RestartMessageCreate(ctx context.Context, message *Message) error {
	err := r.messageCollection.CreateWithCtx(ctx, message)
	if err != nil {
		var mgerr mongo.WriteException
		if !(errors.As(err, &mgerr) && mgerr.HasErrorCode(duplicationError)) {
			return errors.WithStack(err)
		}

		err = r.messageCollection.First(
			bson.M{"tg_chat_id": message.TgChatID, "tg_id": message.TgId},
			message,
		)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = r.restartMessageCollection.CreateWithCtx(ctx, &RestartMessage{
		MessageId: message.ID,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) RestartMessageGetAndDelete(ctx context.Context) ([]Message, error) {
	messages := make([]Message, 0, 5)

	err := r.messageCollection.SimpleAggregateWithCtx(
		ctx,
		&messages,
		builder.Lookup(r.restartMessageCollection.Name(), "_id", "message_id", "restart_messages"),
		bson.M{
			operator.Project: bson.M{
				"text":             1,
				"tg_chat_id":       1,
				"tg_id":            1,
				"created_at":       1,
				"updated_at":       1,
				"restart_messages": "$restart_messages.message_id",
			},
		},
		bson.M{operator.Unwind: "$restart_messages"},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, err = r.restartMessageCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return messages, nil
}

func (r *Repository) PingMessageCreate(
	ctx context.Context,
	message *Message,
	deleteAt time.Time,
) error {
	err := r.messageCollection.CreateWithCtx(ctx, message)
	if err != nil {
		var mgerr mongo.WriteException
		if !(errors.As(err, &mgerr) && mgerr.HasErrorCode(duplicationError)) {
			return errors.WithStack(err)
		}

		err = r.messageCollection.First(
			bson.M{"tg_chat_id": message.TgChatID, "tg_id": message.TgId},
			message,
		)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	err = r.pingMessageCollection.CreateWithCtx(ctx, &PingMessage{
		MessageId: message.ID,
		DeleteAt:  deleteAt,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) PingMessageGetAndDeleteForDeletion(ctx context.Context) ([]Message, error) {
	messages := make([]Message, 0, 5)

	now := time.Now().UTC()

	err := r.messageCollection.SimpleAggregateWithCtx(
		ctx,
		&messages,
		builder.Lookup(r.pingMessageCollection.Name(), "_id", "message_id", "ping_messages"),
		bson.M{operator.Match: bson.M{"ping_messages.delete_at": bson.M{operator.Lt: now}}},
		bson.M{
			operator.Project: bson.M{
				"text":       1,
				"tg_chat_id": 1,
				"tg_id":      1,
				"created_at": 1,
				"updated_at": 1,
			},
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	_, err = r.pingMessageCollection.DeleteMany(ctx, bson.M{"delete_at": bson.M{operator.Lt: now}})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return messages, nil
}
