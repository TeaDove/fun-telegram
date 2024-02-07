package db_repository

import (
	"context"
	"github.com/kamva/mgm/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/goteleout/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func getRepository(t *testing.T) *Repository {
	r, err := New()
	require.NoError(t, err)

	return r
}

func generateMessage(r *Repository, t *testing.T) []Message {
	var wg sync.WaitGroup
	messages := make([]Message, 0, 1000)

	ctx := utils.GetCtx()
	for i := 0; i < rand.Intn(500); i++ {
		for j := 0; j < rand.Intn(500); j++ {
			wg.Add(1)
			chatId := rand.Int63n(1_000_000)

			text := strconv.Itoa(rand.Intn(1_000_000))
			message := Message{Text: text, TgChatID: chatId, TgUserId: rand.Int63n(10), TgId: j}
			messages = append(messages, message)

			go func() {
				defer wg.Done()
				err := r.MessageCreate(ctx, &message)
				require.NoError(t, err)
			}()
		}
	}
	wg.Wait()

	return messages
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

	_, err := r.MessageDeleteOld(utils.GetModuleCtx("repository"))
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

func TestIntegration_DbRepository_CheckUserExists_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.UserUpsert(ctx, &User{
		TgUserId:   id,
		TgUsername: "teadove",
		TgName:     "teadove",
	})
	require.NoError(t, err)

	ok, err := r.CheckUserExists(ctx, id)
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestIntegration_DbRepository_CheckUserExists_False(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)

	ok, err := r.CheckUserExists(ctx, id)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestIntegration_DbRepository_GetUserById_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.UserUpsert(ctx, &User{
		TgUserId:   id,
		TgUsername: "teadove",
		TgName:     "teadove",
	})
	require.NoError(t, err)

	user, err := r.GetUserById(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "teadove", user.TgName)
}

func TestIntegration_DbRepository_GetUsersByChatId_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	_, err := r.GetUsersByChatId(ctx, 1825059942)
	assert.NoError(t, err)
}

func TestIntegration_DbRepository_GetMessagesByChatAndUsername_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	_, err := r.GetMessagesByChatAndUsername(ctx, 1825059942, "TeaDove")
	assert.NoError(t, err)
}

func TestIntegration_DbRepository_DeleteMessagesByChat_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.MessageCreate(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 1})
	require.NoError(t, err)
	err = r.MessageCreate(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 2})
	require.NoError(t, err)
	err = r.MessageCreate(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 3})
	require.NoError(t, err)
	err = r.MessageCreate(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 4})
	require.NoError(t, err)
	err = r.MessageCreate(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 5})
	require.NoError(t, err)

	count, err := r.DeleteMessagesByChat(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), count)
}

func TestIntegration_DbRepository_Ping_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()

	err := r.Ping(ctx)
	assert.NoError(t, err)
}

func TestIntegration_DbRepository_StatsForTable_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()

	_, err := r.StatsForTable(ctx, mgm.CollName(&Message{}))
	assert.NoError(t, err)
}

func TestIntegration_DbRepository_StatsForDatabase_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()

	_, err := r.StatsForDatabase(ctx)
	assert.NoError(t, err)
}

func TestIntegration_DbRepository_MessageGetSortedLimited_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)
	generateMessage(r, t)
	ctx := context.Background()

	messages, err := r.MessageGetSortedLimited(ctx, 1000)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(messages), 0)
	assert.LessOrEqual(t, len(messages), 1000)
}

func TestIntegration_DbRepository_DeleteMessages_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)
	messages := generateMessage(r, t)
	utils.SendInterface(len(messages))

	ctx := context.Background()

	count, err := r.DeleteMessages(ctx, messages)
	assert.NoError(t, err)
	utils.SendInterface(count)
}
