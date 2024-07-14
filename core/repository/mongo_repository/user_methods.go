package mongo_repository

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *Repository) UserUpsert(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now().UTC()

	filter := bson.M{"tg_id": user.TgId}
	update := bson.M{"$set": bson.M{
		"tg_id":       user.TgId,
		"tg_username": user.TgUsername,
		"tg_name":     user.TgName,
		"updated_at":  user.UpdatedAt,
		"created_at":  user.CreatedAt,
		"is_bot":      user.IsBot,
	}}
	opts := options.Update().SetUpsert(true)

	_, err := r.userCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (r *Repository) GetUsersById(ctx context.Context, usersId []int64) ([]User, error) {
	users := make([]User, 0, len(usersId))

	err := r.userCollection.SimpleFindWithCtx(ctx, &users, bson.M{"tg_id": bson.M{"$in": usersId}})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return users, nil
}

func (r *Repository) GetUserById(ctx context.Context, userId int64) (User, error) {
	var user User

	err := r.userCollection.FirstWithCtx(ctx, bson.M{"tg_id": userId}, &user)
	if err != nil {
		return User{}, errors.WithStack(err)
	}

	return user, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (User, error) {
	var user User

	err := r.userCollection.FirstWithCtx(
		ctx,
		bson.M{"tg_username": strings.ToLower(username)},
		&user,
	)
	if err != nil {
		return User{}, errors.WithStack(err)
	}

	return user, nil
}
