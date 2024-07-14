package mongo_repository

import (
	"context"
	"time"

	"github.com/kamva/mgm/v3/builder"
	"github.com/kamva/mgm/v3/operator"
	errors "github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const duplicationError = 11000

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
