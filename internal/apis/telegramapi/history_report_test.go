package telegramapi

import (
	"testing"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/stretchr/testify/require"
)

func TestHistoryDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		span time.Duration
		want string
	}{
		{span: 45 * historyDay, want: "1.5M"},
		{span: 74*time.Hour + 24*time.Minute, want: "3.1d"},
		{span: 270 * historyDay, want: "9M"},
		{span: time.Minute, want: "1m"},
		{span: 90 * time.Minute, want: "1.5h"},
		{span: 65 * time.Second, want: "1.1m"},
		{span: 0, want: "0s"},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, formatHistoryDuration(test.span))
		})
	}
}

func TestHistoryCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		amount int
		want   string
	}{
		{amount: 0, want: "0"},
		{amount: 1, want: "1"},
		{amount: 999, want: "999"},
		{amount: 1000, want: "1k"},
		{amount: 1256, want: "1.3k"},
		{amount: 10000, want: "10k"},
		{amount: 12345, want: "12.3k"},
		{amount: 999900, want: "999.9k"},
		{amount: 999999, want: "1kk"},
		{amount: 1000000, want: "1kk"},
		{amount: 1234567, want: "1.2kk"},
		{amount: 999999999, want: "1kkk"},
		{amount: 1234567890, want: "1.2kkk"},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, test.want, formatHistoryCount(test.amount))
		})
	}
}

func TestHistoryReportGroupsPeers(t *testing.T) {
	t.Parallel()

	const start = int64(1_000_000)

	chats := historyChats{
		{id: -2, name: "A"}, {id: -3, name: "A"}, {id: -4, name: "B"}, {id: -5, name: "Empty"},
	}
	summary := historySummary{}
	summary.append(tgdtos.Message{PeerID: -2, Date: start})
	summary.append(tgdtos.Message{PeerID: -3, Date: start + int64((45*historyDay)/time.Second)})
	summary.append(tgdtos.Message{PeerID: -4, Date: start + int64(60*historyDay/time.Second)})
	summary.append(tgdtos.Message{PeerID: -4, Date: start + int64(60*historyDay/time.Second) + 60})

	builder := entity.Builder{}
	err := styling.Perform(&builder, summary.styledCaption(chats)...)
	require.NoError(t, err)

	caption, _ := builder.Complete()

	require.Equal(t,
		"A, msgs=2, range=1.5M\nB, msgs=2, range=1m\nEmpty, msgs=0, range=0s\n\nTotal, msgs=4, range=2M",
		caption,
	)
}
