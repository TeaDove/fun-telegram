package telegramapi

import (
	"context"
	"testing"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/cockroachdb/errors"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestMergedHistoryLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		count  int
		cutoff int64
		peers  []int64
		stop   string
	}{
		{name: "newest across peers", count: 2, cutoff: 1, peers: []int64{-2, -3}, stop: "message limit"},
		{name: "date across peers", count: 10, cutoff: 20, peers: []int64{-2, -3, -2}, stop: "date limit"},
		{name: "all histories", count: 10, cutoff: 1, peers: []int64{-2, -3, -2, -3}, stop: "end of history"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			sources := []historySource{
				{chat: historyChat{id: -2}, iter: &testHistoryIterator{records: []tg.NotEmptyMessage{
					&tg.Message{ID: 2, Date: 40}, &tg.Message{ID: 1, Date: 20},
				}}},
				{chat: historyChat{id: -3}, iter: &testHistoryIterator{records: []tg.NotEmptyMessage{
					&tg.Message{ID: 2, Date: 30}, &tg.Message{ID: 1, Date: 10},
				}}},
			}
			records := make([]tgdtos.Message, 0)

			summary, err := walkHistory(t.Context(), sources, historyOptions{
				count: test.count, cutoff: time.Unix(test.cutoff, 0),
			}, 1, func(record tgdtos.Message, _ tgdtos.Author) error {
				records = append(records, record)
				return nil
			}, func(context.Context, string) {})
			require.NoError(t, err)
			require.Equal(t, test.stop, summary.stop)
			require.Equal(t, len(test.peers), summary.count)

			peers := make([]int64, 0, len(records))
			for _, record := range records {
				peers = append(peers, record.PeerID)
			}

			require.Equal(t, test.peers, peers)
			require.Equal(t, records[0].ID, records[1].ID)
		})
	}
}

func TestMergedHistorySourceFailure(t *testing.T) {
	t.Parallel()

	failure := errors.New("history unavailable")
	sources := []historySource{
		{chat: historyChat{id: -2}, iter: &testHistoryIterator{records: []tg.NotEmptyMessage{
			&tg.Message{ID: 1, Date: 20},
		}}},
		{chat: historyChat{id: -3}, iter: &testHistoryIterator{err: failure}},
	}
	consumed := 0
	_, err := walkHistory(t.Context(), sources, historyOptions{count: 10}, 1,
		func(tgdtos.Message, tgdtos.Author) error {
			consumed++
			return nil
		}, func(context.Context, string) {})
	require.ErrorIs(t, err, failure)
	require.Zero(t, consumed)
}
