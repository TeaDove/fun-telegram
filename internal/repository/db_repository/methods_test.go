package db_repository

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/goteleout/internal/shared"
	"github.com/teadove/goteleout/internal/utils"
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
	t.Parallel()
	r := getRepository(t)

	id := rand.Int63n(1_000_000)
	err := r.MessageCreate(context.Background(), &Message{Text: "Привет", TgChatID: id, TgUserId: id, TgId: 1})
	assert.NoError(t, err)

	message := Message{}

	err = r.messageCollection.First(bson.M{"tg_chat_id": id, "tg_user_id": id}, &message)
	assert.NoError(t, err)

	assert.Equal(t, "Привет", message.Text)
}

func TestIntegration_DbRepository_UserUpsert_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	id := rand.Int63n(1_000_000)
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

func TestIntegration_DbRepository_DeleteOldMessages_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	err := r.MessageDeleteOld(utils.GetModuleCtx("repository"))
	assert.NoError(t, err)
}

func TestIntegration_DbRepository_GetMessagesByChat_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	id := rand.Int63n(1_000_000)
	ctx := context.Background()
	err := r.MessageCreate(ctx, &Message{Text: "Привет", TgChatID: id, TgUserId: id, TgId: 1})
	assert.NoError(t, err)
	err = r.MessageCreate(ctx, &Message{Text: "Привет", TgChatID: id, TgUserId: id, TgId: 2})
	assert.NoError(t, err)

	messages, err := r.GetMessagesByChat(ctx, id)
	assert.NoError(t, err)

	assert.Len(t, messages, 2)
	assert.Equal(t, "Привет", messages[0].Text)
}

func TestIntegration_DbRepository_GetUsersByUserId_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.UserUpsert(context.Background(), &User{
		TgUserId:   id,
		TgUsername: "teadove",
		TgName:     "teadove",
	})
	assert.NoError(t, err)

	users, err := r.GetUsersById(ctx, []int64{id})
	assert.NoError(t, err)

	assert.Len(t, users, 1)
	assert.Equal(t, "teadove", users[0].TgUsername)
}

func TestIntegration_DbRepository_GetLastMessage_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.MessageCreate(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 1})
	require.NoError(t, err)
	err = r.MessageCreate(ctx, &Message{Text: "2", TgChatID: id, TgUserId: id, TgId: 2})
	require.NoError(t, err)

	msg, err := r.GetLastMessage(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, "1", msg.Text)
}

func TestIntegration_DbRepository_MessageCreateOrNothingAndSetTime_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.MessageCreateOrNothingAndSetTime(ctx, &Message{Text: "2", TgChatID: id, TgUserId: id, TgId: 100})
	require.NoError(t, err)
	err = r.MessageCreateOrNothingAndSetTime(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 100})
	require.NoError(t, err)

	msg, err := r.GetLastMessage(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, "2", msg.Text)
}
