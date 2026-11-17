package telegramapi

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestDumpChatRecords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		migrated bool
		oldTitle string
		onlyOld  bool
		want     []tgdtos.Chat
	}{
		{
			name: "independent chats with the same name", oldTitle: "A",
			want: []tgdtos.Chat{
				{Title: "A", PeerIDs: []int64{-2}},
				{Title: "A", PeerIDs: []int64{-1000000000003}},
				{Title: "A", PeerIDs: []int64{-4}},
			},
		},
		{
			name: "migration with another same-name chat", migrated: true, oldTitle: "A",
			want: []tgdtos.Chat{
				{Title: "A", PeerIDs: []int64{-2, -1000000000003}},
				{Title: "A", PeerIDs: []int64{-4}},
			},
		},
		{
			name: "renamed migrated chat", migrated: true, oldTitle: "Old title",
			want: []tgdtos.Chat{
				{Title: "A", PeerIDs: []int64{-2, -1000000000003}},
				{Title: "A", PeerIDs: []int64{-4}},
			},
		},
		{
			name: "selected original peer only", migrated: true, oldTitle: "Old title", onlyOld: true,
			want: []tgdtos.Chat{{Title: "Old title", PeerIDs: []int64{-2}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			old := &tg.Chat{ID: 2, Title: test.oldTitle}
			if test.migrated {
				old.MigratedTo = &tg.InputChannel{ChannelID: 3}
			}

			current := &tg.Channel{ID: 3, Title: "A"}
			entities := peer.NewEntities(
				nil,
				map[int64]*tg.Chat{old.ID: old},
				map[int64]*tg.Channel{current.ID: current},
			)

			chats := historyChats{
				historyChatFromPeer(&tg.PeerChat{ChatID: old.ID}, &tg.InputPeerChat{ChatID: old.ID}, entities),
			}
			if !test.onlyOld {
				chats = append(chats,
					historyChatFromPeer(&tg.PeerChannel{ChannelID: current.ID},
						&tg.InputPeerChannel{ChannelID: current.ID}, entities),
					historyChat{id: -4, name: "A"},
				)
			}

			var output bytes.Buffer

			_, err := newHistoryDump(&output, chats)
			require.NoError(t, err)

			decoder := json.NewDecoder(&output)

			for _, expected := range test.want {
				var raw json.RawMessage
				require.NoError(t, decoder.Decode(&raw))
				require.NotContains(t, string(raw), "version")

				var record tgdtos.Record[tgdtos.Chat]
				require.NoError(t, json.Unmarshal(raw, &record))
				require.Equal(t, tgdtos.TGTypeChat, record.Type)
				require.Equal(t, expected, record.Payload)
			}

			var extra json.RawMessage
			require.ErrorIs(t, decoder.Decode(&extra), io.EOF)
		})
	}
}
