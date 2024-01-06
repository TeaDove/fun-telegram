package user_repository

import (
	"context"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

func (r *Repository) CreateUser(ctx context.Context, user *User) (*User, error) {
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Time{}

	res, err := r.userCollection.InsertOne(ctx, &user)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("res.InsertedID is not primitive.ObjectID")
	}

	user.ID = id

	return user, nil
}

func (r *Repository) GetUser(ctx context.Context, tgUserId int64) (*User, error) {
	var user User

	err := r.userCollection.FindOne(ctx, bson.D{{Key: "tg_user_id", Value: tgUserId}}).Decode(&user)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &user, nil
}

func (r *Repository) CreateUserInChat(ctx context.Context, userInChat *UserInChat) (*UserInChat, error) {
	userInChat.CreatedAt = time.Now().UTC()
	userInChat.UpdatedAt = time.Time{}

	res, err := r.userInChatCollection.InsertOne(ctx, &userInChat)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("res.InsertedID is not primitive.ObjectID")
	}

	userInChat.ID = id

	return userInChat, nil
}
