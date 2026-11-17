package telegramapi

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
	"github.com/stretchr/testify/require"
)

func TestResolveMigratedHistoryChats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		command   string
		oldListed bool
		saved     bool
	}{
		{name: "same title", command: `!stats "Current title"`, oldListed: true, saved: true},
		{name: "old title", command: `!stats "Old title"`, oldListed: true, saved: true},
		{name: "current peer", command: "!stats --peer-id=-1000000000003", saved: true},
		{name: "hidden old peer", command: "!stats --peer-id=-2", saved: true},
		{
			name: "repeated migration peers", oldListed: true, saved: true,
			command: "!stats --peer-id=-2 --peer-id=-1000000000003 --peer-id=-2",
		},
		{name: "current chat", command: "!stats"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			channel := &tg.Channel{ID: 3, Title: "Current title", Forum: true, Photo: &tg.ChatPhotoEmpty{}}

			oldTitle := channel.Title
			if test.name == "old title" {
				oldTitle = "Old title"
			}

			old := &tg.Chat{
				ID: 2, Title: oldTitle, MigratedTo: &tg.InputChannel{ChannelID: channel.ID}, Photo: &tg.ChatPhotoEmpty{},
			}
			fullCalls := 0
			api := newMigrationTestAPI(t, channel, old, test.oldListed, &fullCalls)
			c := getOpt(test.command, historyFlags()...)
			c.extCtx = &ext.Context{Context: t.Context(), Self: &tg.User{ID: 1}}

			origin := tg.PeerClass(&tg.PeerChannel{ChannelID: channel.ID})
			if test.saved {
				origin = &tg.PeerUser{UserID: c.extCtx.Self.ID}
			}

			c.update = &ext.Update{
				EffectiveMessage: types.ConstructMessage(&tg.Message{PeerID: origin}),
				Entities:         &tg.Entities{Channels: map[int64]*tg.Channel{channel.ID: channel}},
			}
			presentation := Presentation{telegramAPI: api}

			chats, err := presentation.resolveHistoryChats(&c)
			require.NoError(t, err)
			require.Equal(t, 1, fullCalls)
			require.Len(t, chats, 2)
			require.Equal(t, constant.TDLibPeerID(-1000000000003), chats[0].id)
			require.Equal(t, constant.TDLibPeerID(-2), chats[1].id)
			require.Equal(t, chats[0].id, chats[1].canonicalID())
			require.True(t, chats[0].forum)
			require.False(t, chats[1].forum)

			var output bytes.Buffer

			_, err = newHistoryDump(&output, chats)
			require.NoError(t, err)

			decoder := json.NewDecoder(&output)

			var header tgdtos.Record[tgdtos.HistoryHeader]
			require.NoError(t, decoder.Decode(&header))
			require.NoError(t, decoder.Decode(&header))
			require.Equal(t, int64(chats[1].id), header.Payload.PeerID)
			require.Equal(t, int64(chats[0].id), header.Payload.MigratedToPeerID)
		})
	}
}

func TestHistoryRepeatedPeerIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		command string
		want    []historyChatSelector
		hasErr  bool
	}{
		{
			command: "!stats --peer-id=-2 --peer-id=-3 -c=1",
			want:    []historyChatSelector{{id: -2}, {id: -3}},
		},
		{command: "!stats --peer-id=-2 --peer-id=invalid --peer-id=-3", hasErr: true},
		{command: "!stats --peer-id=-2 --peer-id=", hasErr: true},
		{command: "!stats --peer-id=-2 --peer-id=0", hasErr: true},
		{command: `!stats "Title" --peer-id=-2`, hasErr: true},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			t.Parallel()

			c := getOpt(test.command, historyFlags()...)

			selectors, err := parseHistoryChatSelectors(c.Text, c.Ops, true)
			if test.hasErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.want, selectors)
		})
	}
}

func newMigrationTestAPI(
	t *testing.T,
	channel *tg.Channel,
	old *tg.Chat,
	oldListed bool,
	fullCalls *int,
) *tg.Client {
	t.Helper()

	return tg.NewClient(tgmock.Invoker(func(request bin.Encoder) (bin.Encoder, error) {
		switch request := request.(type) {
		case *tg.MessagesGetDialogsRequest:
			if request.FolderID != 0 {
				return &tg.MessagesDialogs{}, nil
			}

			dialogs := &tg.MessagesDialogs{
				Chats:   []tg.ChatClass{channel},
				Dialogs: []tg.DialogClass{&tg.Dialog{Peer: &tg.PeerChannel{ChannelID: channel.ID}}},
			}
			if oldListed {
				dialogs.Chats = append(dialogs.Chats, old)
				dialogs.Dialogs = append(dialogs.Dialogs, &tg.Dialog{Peer: &tg.PeerChat{ChatID: old.ID}})
			}

			return dialogs, nil
		case *tg.MessagesGetChatsRequest:
			require.Equal(t, []int64{old.ID}, request.ID)
			return &tg.MessagesChats{Chats: []tg.ChatClass{old}}, nil
		case *tg.ChannelsGetFullChannelRequest:
			require.Equal(t, &tg.InputChannel{ChannelID: channel.ID}, request.Channel)

			(*fullCalls)++

			return &tg.MessagesChatFull{
				FullChat: &tg.ChannelFull{
					ID: channel.ID, MigratedFromChatID: old.ID, ChatPhoto: &tg.PhotoEmpty{},
				},
				Chats: []tg.ChatClass{channel},
			}, nil
		default:
			t.Fatalf("unexpected Telegram request %T", request)
			return nil, nil
		}
	}))
}
