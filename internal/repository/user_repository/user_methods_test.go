package user_repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
)

var ctx = context.Background()

func getRepository(t *testing.T) *Repository {
	conn, err := MongoDBConnect(ctx, "mongodb://localhost:27017")
	assert.NoError(t, err)

	r, err := New(conn)
	assert.NoError(t, err)

	return r
}

func TestIntegration_UserRepository_CreateUser_Ok(t *testing.T) {
	r := getRepository(t)

	_, err := r.CreateUser(ctx, &User{
		TgUserID:   123,
		TgUsername: "TeaDove",
	})
	assert.NoError(t, err)
}

func TestIntegration_UserRepository_GetUser_Ok(t *testing.T) {
	r := getRepository(t)

	_, err := r.CreateUser(ctx, &User{
		TgUserID:   123,
		TgUsername: "TeaDove",
	})
	assert.NoError(t, err)

	user, err := r.GetUser(ctx, 123)
	assert.NoError(t, err)

	assert.Equal(t, "TeaDove", user.TgUsername)
	assert.Equal(t, int64(123), user.TgUserID)
}

func TestIntegration_UserRepository_CreateUserInChat_Ok(t *testing.T) {
	r := getRepository(t)

	_, err := r.CreateUserInChat(ctx, &UserInChat{
		TgChatID: 123,
		UserID:   primitive.NewObjectID(),
		Toxicity: 0.1,
	})
	assert.NoError(t, err)
}
