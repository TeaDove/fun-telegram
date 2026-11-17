package telegramapi

import (
	"testing"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
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

func TestHistoryReportGroupsPeers(t *testing.T) {
	t.Parallel()

	const start = int64(1_000_000)

	chats := historyChats{
		{id: -2, name: "A"}, {id: -3, name: "A"}, {id: -4, name: "B"}, {id: -5, name: "Empty"},
	}
	summary := historySummary{}
	summary.append(tgdtos.Message{PeerID: -2, Date: start})
	summary.append(tgdtos.Message{PeerID: -3, Date: start + int64((45*historyDay)/time.Second)})
	summary.append(tgdtos.Message{PeerID: -4, Date: start})
	summary.append(tgdtos.Message{PeerID: -4, Date: start + 60})

	require.Equal(t,
		"A (-2, -3), msgs=2, range=1.5M\nB (-4), msgs=2, range=1m\nEmpty (-5), msgs=0, range=0s",
		summary.caption(chats),
	)
}
