package db_repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/teadove/goteleout/internal/shared"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"testing"
)

func getRepository(t *testing.T) *Repository {
	r, err := New(shared.AppSettings.Storage.MongoDbUrl)
	assert.NoError(t, err)

	return r
}

func TestIntegration_DbRepository_MessageCreate_Ok(t *testing.T) {
	r := getRepository(t)

	err := r.MessageCreate(context.Background(), &Message{Text: "Привет", TgChatID: 123, TgUserId: 123})
	assert.NoError(t, err)

	message := Message{}

	err = r.messageCollection.First(bson.M{"tg_chat_id": 123, "tg_user_id": 123}, &message)
	assert.NoError(t, err)

	assert.Equal(t, "Привет", message.Text)
}

func TestIntegration_DbRepository_UserUpsert_Ok(t *testing.T) {
	r := getRepository(t)

	id := rand.Int63n(2000)
	err := r.UserUpsert(context.Background(), &User{
		TgUserId:   id,
		TgUsername: "teadove",
		TgName:     "teadove",
	})
	assert.NoError(t, err)

	err = r.UserUpsert(context.Background(), &User{
		TgUserId:   id,
		TgUsername: "tainella",
		TgName:     "tainella",
	})
	assert.NoError(t, err)
}
