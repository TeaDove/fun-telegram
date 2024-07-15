package db_repository

import (
	"math/rand/v2"
	"testing"

	"github.com/guregu/null/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teadove/fun_telegram/core/infrastructure/pg"
	"github.com/teadove/fun_telegram/core/shared"
)

func generateMessage() Message {
	return Message{
		TgChatID: rand.Int64N(100000),
		TgId:     rand.IntN(100000),
		TgUserId: rand.Int64N(100000),
		Text:     shared.RandomString(),
	}
}

func TestIntegration_DbRepository_MessageCountByChatIdAndUserId_Ok(t *testing.T) {
	ctx := shared.GetCtx()
	db, err := pg.NewClientFromSettings()
	require.NoError(t, err)

	dbRepository, err := NewRepository(ctx, db)
	require.NoError(t, err)

	chatId := rand.Int64N(100_000)
	userId := rand.Int64N(100_000)

	message := generateMessage()
	message.TgChatID = chatId
	message.TgUserId = userId
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	message = generateMessage()
	message.TgChatID = chatId
	message.TgUserId = userId
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	count, err := dbRepository.MessageCountByChatIdAndUserId(ctx, chatId, userId)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), count)
}

func TestIntegration_DbRepository_MessageGroupByChatIdAndUserId_Ok(t *testing.T) {
	ctx := shared.GetCtx()
	db, err := pg.NewClientFromSettings()
	require.NoError(t, err)

	dbRepository, err := NewRepository(ctx, db)
	require.NoError(t, err)

	chatId := rand.Int64N(100_000)
	userId := rand.Int64N(100_000)

	message := generateMessage()
	message.TgChatID = chatId
	message.TgUserId = userId
	message.ToxicWordsCount = 3
	message.WordsCount = 5
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	message = generateMessage()
	message.TgChatID = chatId
	message.TgUserId = userId
	message.ToxicWordsCount = 2
	message.WordsCount = 4
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	group, err := dbRepository.MessageGroupByChatIdAndUserId(ctx, chatId, []int64{userId}, 10, true)
	require.NoError(t, err)

	assert.Equal(t, uint64(9), group[0].WordsCount)
	assert.Equal(t, uint64(5), group[0].ToxicWordsCount)
	assert.Equal(t, userId, group[0].TgUserId)
}

func TestIntegration_DbRepository_MessageInsert_Ok(t *testing.T) {
	ctx := shared.GetCtx()
	db, err := pg.NewClientFromSettings()
	require.NoError(t, err)

	dbRepository, err := NewRepository(ctx, db)
	require.NoError(t, err)

	message := generateMessage()
	chatId := rand.Int64N(100_000)
	userId := rand.Int64N(100_000)
	msgId := message.TgId

	message.TgChatID = chatId
	message.TgUserId = userId
	message.ToxicWordsCount = 3
	message.WordsCount = 5
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	message = generateMessage()
	message.TgChatID = chatId
	message.TgUserId = userId
	message.ReplyToTgMsgID = null.IntFrom(int64(msgId))
	message.ToxicWordsCount = 3
	message.WordsCount = 5
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)
}
