package mongo_repository

import (
	"context"
	"time"

	"github.com/kamva/mgm/v3/builder"
	"github.com/kamva/mgm/v3/operator"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *Repository) MemberUpsert(ctx context.Context, member *Member) error {
	member.CreatedAt = time.Now().UTC()
	member.UpdatedAt = time.Now().UTC()

	filter := bson.M{"tg_chat_id": member.TgChatId, "tg_user_id": member.TgUserId}
	update := bson.M{"$set": bson.M{
		"tg_user_id": member.TgUserId,
		"tg_chat_id": member.TgChatId,
		"status":     member.Status,
		"updated_at": member.UpdatedAt,
		"created_at": member.CreatedAt,
	}}
	opts := options.Update().SetUpsert(true)

	_, err := r.memberCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) GetUsersInChat(ctx context.Context, chatId int64) (UsersInChat, error) {
	usersInChat := make(UsersInChat, 0, 100)

	err := r.memberCollection.SimpleAggregateWithCtx(
		ctx,
		&usersInChat,
		builder.Lookup(r.userCollection.Name(), "tg_user_id", "tg_id", "user"),
		bson.M{operator.Match: bson.M{"tg_chat_id": chatId}},
		bson.M{operator.Unwind: "$user"},
		bson.M{
			operator.Project: bson.M{
				"status":      1,
				"tg_id":       "$user.tg_id",
				"tg_username": "$user.tg_username",
				"tg_name":     "$user.tg_name",
				"is_bot":      "$user.is_bot",
				// "tg_chat_id":  1,
			},
		},
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return usersInChat, nil
}
