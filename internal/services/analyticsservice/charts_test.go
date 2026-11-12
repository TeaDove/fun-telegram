package analyticsservice_test

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"testing"
	"time"

	"github.com/TeaDove/fun-telegram/internal/services/analyticsservice"
	"github.com/TeaDove/fun-telegram/internal/services/messageservice"
	"github.com/stretchr/testify/require"
)

func TestAnaliseChatJPEG(t *testing.T) {
	t.Parallel()

	service, err := analyticsservice.New()
	require.NoError(t, err)

	const participantCount = 35

	start := time.Date(2026, time.January, 5, 0, 0, 0, 0, time.UTC)

	storage := messageservice.Storage{
		Messages: make(messageservice.Messages, 0, participantCount),
		Users:    make(messageservice.UsersInChat, 0, participantCount),
	}
	for index := range participantCount {
		id := int64(index + 1)
		storage.Messages = append(storage.Messages, messageservice.Message{
			TgUserID:        id,
			CreatedAt:       start.AddDate(0, 0, index*7),
			WordsCount:      uint64(index * 37),
			ToxicWordsCount: uint64(index / 3),
		})
		storage.Users = append(storage.Users, messageservice.UserInChat{
			TgID:       id,
			TgName:     fmt.Sprintf("Участник с длинной подписью %d", index+1),
			TgUsername: fmt.Sprintf("sample_%d", index+1),
		})
	}

	storage.UsersNameGetter = storage.Users.GetNameGetter()
	report, err := service.AnaliseChat(t.Context(), &analyticsservice.AnaliseChatInput{Storage: storage})
	require.NoError(t, err)
	require.Equal(t, participantCount, report.MessagesCount)

	expected := []string{
		"AntiChatterBoxes.jpeg", "ChatterBoxes.jpeg", "MessagesGroupedByDateByChatId.jpeg", "MostToxicUsers.jpeg",
	}
	require.Len(t, report.Images, len(expected))

	for index, chart := range report.Images {
		require.Equal(t, expected[index], chart.Filename())

		config, err := jpeg.DecodeConfig(bytes.NewReader(chart.Content))
		require.NoError(t, err)
		require.GreaterOrEqual(t, config.Width, 4950)
		require.GreaterOrEqual(t, config.Height, 3000)
		require.LessOrEqual(t, config.Width+config.Height, 10000)
		require.Equal(t, "JFIF", string(chart.Content[6:10]))
		require.Equal(t, []byte{1, 1, 44, 1, 44}, chart.Content[13:18])
	}
}
