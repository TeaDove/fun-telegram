package db_repository

import (
	"math/rand/v2"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/guregu/null/v5"
	"github.com/teadove/teasutils/utils/random_utils"
	"github.com/teadove/teasutils/utils/test_utils"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateMessage() Message {
	return Message{
		TgChatID: rand.Int64N(100000),
		TgID:     rand.IntN(100000),
		TgUserID: rand.Int64N(100000),
		Text:     random_utils.Text(),
	}
}

func getDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)

	return db
}

func TestIntegration_DbRepository_MessageCountByChatIdAndUserId_Ok(t *testing.T) {
	t.Parallel()

	ctx := test_utils.GetLoggedContext()
	db := getDB(t)

	dbRepository, err := NewRepository(ctx, db)
	require.NoError(t, err)

	chatID := rand.Int64N(100_000)
	userID := rand.Int64N(100_000)

	message := generateMessage()
	message.TgChatID = chatID
	message.TgUserID = userID
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	message = generateMessage()
	message.TgChatID = chatID
	message.TgUserID = userID
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	count, err := dbRepository.MessageCountByChatIDAndUserID(ctx, chatID, userID)
	require.NoError(t, err)
	assert.Equal(t, uint64(2), count)
}

func TestIntegration_DbRepository_MessageGroupByChatIdAndUserId_Ok(t *testing.T) {
	t.Parallel()

	ctx := test_utils.GetLoggedContext()
	db := getDB(t)

	dbRepository, err := NewRepository(ctx, db)
	require.NoError(t, err)

	chatID := rand.Int64N(100_000)
	userID := rand.Int64N(100_000)

	message := generateMessage()
	message.TgChatID = chatID
	message.TgUserID = userID
	message.ToxicWordsCount = 3
	message.WordsCount = 5
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	message = generateMessage()
	message.TgChatID = chatID
	message.TgUserID = userID
	message.ToxicWordsCount = 2
	message.WordsCount = 4
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	group, err := dbRepository.MessageGroupByChatIDAndUserID(ctx, chatID, []int64{userID}, 10, true)
	require.NoError(t, err)

	assert.Equal(t, uint64(9), group[0].WordsCount)
	assert.Equal(t, uint64(5), group[0].ToxicWordsCount)
	assert.Equal(t, userID, group[0].TgUserID)
}

func TestIntegration_DbRepository_MessageInsert_Ok(t *testing.T) {
	t.Parallel()

	ctx := test_utils.GetLoggedContext()
	db := getDB(t)

	dbRepository, err := NewRepository(ctx, db)
	require.NoError(t, err)

	message := generateMessage()
	chatID := rand.Int64N(100_000)
	userID := rand.Int64N(100_000)
	msgID := message.TgID

	message.TgChatID = chatID
	message.TgUserID = userID
	message.ToxicWordsCount = 3
	message.WordsCount = 5
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)

	message = generateMessage()
	message.TgChatID = chatID
	message.TgUserID = userID
	message.ReplyToTgMsgID = null.IntFrom(int64(msgID))
	message.ToxicWordsCount = 3
	message.WordsCount = 5
	err = dbRepository.MessageInsert(ctx, &message)
	require.NoError(t, err)
}
