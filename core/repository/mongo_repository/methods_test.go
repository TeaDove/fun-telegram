package mongo_repository

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/kamva/mgm/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/fun_telegram/core/shared"
)

func getRepository(t *testing.T) *Repository {
	r, err := New()
	require.NoError(t, err)

	return r
}

func TestIntegration_DbRepository_UserUpsert_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)
	ctx := shared.GetCtx()

	id := rand.Int63n(1_000_000)
	err := r.UserUpsert(ctx, &User{
		TgId:       id,
		TgUsername: "teadove",
		TgName:     "teadove",
	})
	assert.NoError(t, err)

	user, err := r.GetUserById(ctx, id)
	assert.Equal(t, user.TgName, "teadove")
	assert.Equal(t, user.TgUsername, "teadove")

	err = r.UserUpsert(ctx, &User{
		TgId:       id,
		TgUsername: "tainella",
		TgName:     "tainella",
	})
	assert.NoError(t, err)

	user, err = r.GetUserById(ctx, id)
	assert.Equal(t, user.TgName, "tainella")
	assert.Equal(t, user.TgUsername, "tainella")
}

func TestIntegration_DbRepository_GetUsersByUserId_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.UserUpsert(context.Background(), &User{
		TgId:       id,
		TgUsername: "teadove",
		TgName:     "teadove",
	})
	assert.NoError(t, err)

	users, err := r.GetUsersById(ctx, []int64{id})
	assert.NoError(t, err)

	assert.Len(t, users, 1)
	assert.Equal(t, "teadove", users[0].TgUsername)
}

func TestIntegration_DbRepository_GetUserById_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.UserUpsert(ctx, &User{
		TgId:       id,
		TgUsername: "teadove",
		TgName:     "teadove",
	})
	require.NoError(t, err)

	user, err := r.GetUserById(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, "teadove", user.TgName)
}

func TestIntegration_DbRepository_Ping_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()

	err := r.Ping(ctx)
	assert.NoError(t, err)
}

func TestIntegration_DbRepository_StatsForTable_Ok(t *testing.T) {
	r := getRepository(t)
	generateMessage(r, t)
	ctx := context.Background()

	_, err := r.StatsForTable(ctx, mgm.CollName(&Message{}))
	assert.NoError(t, err)
}

func TestIntegration_DbRepository_StatsForDatabase_Ok(t *testing.T) {
	r := getRepository(t)
	generateMessage(r, t)
	ctx := context.Background()

	stats, err := r.StatsForDatabase(ctx)
	assert.NoError(t, err)
	shared.SendInterface(stats)
}

func TestIntegration_DbRepository_ReloadMessageCreateGetAndDelete_Ok(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.RestartMessageCreate(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 1})
	require.NoError(t, err)

	id = rand.Int63n(1_000_000)
	err = r.RestartMessageCreate(ctx, &Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 1})
	require.NoError(t, err)

	_, err = r.RestartMessageGetAndDelete(ctx)
	assert.NoError(t, err)

	messages, err := r.RestartMessageGetAndDelete(ctx)
	assert.NoError(t, err)
	assert.Len(t, messages, 0)
}

func TestIntegration_DbRepository_PingMessageGetAndDeleteForDeletion_NoMessagesOk(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.PingMessageCreate(
		ctx,
		&Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 1},
		time.Now().UTC().Add(1*time.Hour),
	)
	require.NoError(t, err)

	messages, err := r.PingMessageGetAndDeleteForDeletion(ctx)
	assert.NoError(t, err)
	assert.Len(t, messages, 0)
}

func TestIntegration_DbRepository_PingMessageGetAndDeleteForDeletion_OneMessageOk(t *testing.T) {
	t.Parallel()
	r := getRepository(t)

	ctx := context.Background()
	id := rand.Int63n(1_000_000)
	err := r.PingMessageCreate(
		ctx,
		&Message{Text: "1", TgChatID: id, TgUserId: id, TgId: 1},
		time.Now().UTC().Add(-1*time.Hour),
	)
	require.NoError(t, err)

	messages, err := r.PingMessageGetAndDeleteForDeletion(ctx)
	assert.NoError(t, err)
	assert.Len(t, messages, 1)
}
