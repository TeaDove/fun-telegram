package telegramapi

import (
	"testing"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
	"github.com/stretchr/testify/require"
)

func TestSelectHistoryNames(t *testing.T) {
	t.Parallel()

	catalog := map[constant.TDLibPeerID]historyChat{
		-2:             {id: -2, name: "A"},
		-1000000000003: {id: -1000000000003, name: "A"},
		-4:             {id: -4, name: "B"},
	}
	tests := []struct {
		command string
		want    []constant.TDLibPeerID
		errText string
	}{
		{command: "!stats -n A -n B -n A", want: []constant.TDLibPeerID{-1000000000003, -2, -4}},
		{command: "!stats -n a", want: []constant.TDLibPeerID{-1000000000003, -2}},
		{command: "!stats --name=A --name=B", want: []constant.TDLibPeerID{-1000000000003, -2, -4}},
		{command: `!stats "A"`, want: []constant.TDLibPeerID{-1000000000003, -2}},
		{command: "!stats -n missing", errText: `chat "missing" not found`},
		{command: "!stats -n A -n missing", errText: `chat "missing" not found`},
		{command: "!stats -n", errText: "specify a chat name after -n"},
		{command: "!stats -n -q", errText: "specify a chat name after -n"},
		{command: "!stats --name=", errText: "specify a chat name after -n"},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			t.Parallel()

			c := getOpt(test.command, historyFlags()...)

			names, err := historyChatNames(&c)
			if err != nil {
				require.ErrorContains(t, err, test.errText)
				return
			}

			chats, err := selectHistoryChats(names, catalog)
			if test.errText != "" {
				require.ErrorContains(t, err, test.errText)
				return
			}

			require.NoError(t, err)

			ids := make([]constant.TDLibPeerID, 0, len(chats))
			for _, chat := range chats {
				ids = append(ids, chat.id)
			}

			require.Equal(t, test.want, ids)
		})
	}
}

func TestSelectSameNameWithoutMigration(t *testing.T) {
	t.Parallel()

	peer := &tg.PeerChat{ChatID: 2}
	group := &tg.Chat{ID: 2, Title: "A", Photo: &tg.ChatPhotoEmpty{}}
	channel := &tg.Channel{ID: 3, Title: "A", Photo: &tg.ChatPhotoEmpty{}}
	mock := tgmock.NewRequire(t)
	mock.Expect().ThenResult(&tg.UpdateShortSentMessage{ID: 1})
	mock.Expect().ThenResult(&tg.MessagesDialogs{
		Dialogs: []tg.DialogClass{
			&tg.Dialog{Peer: peer}, &tg.Dialog{Peer: &tg.PeerChannel{ChannelID: channel.ID}},
		},
		Chats: []tg.ChatClass{group, channel},
	})
	mock.Expect().ThenResult(&tg.MessagesDialogs{})
	mock.Expect().ThenResult(&tg.MessagesDialogs{})
	api := tg.NewClient(mock)
	r := Presentation{telegramAPI: api}
	c := Context{
		Silent: true,
		extCtx: ext.NewContext(t.Context(), api, nil, &tg.User{ID: 1}, message.NewSender(api), nil, false),
		update: &ext.Update{
			EffectiveMessage: types.ConstructMessage(&tg.Message{PeerID: peer}),
			Entities:         &tg.Entities{Chats: map[int64]*tg.Chat{group.ID: group}},
		},
	}
	request, err := r.prepareHistory(&c)
	require.NoError(t, err)
	require.Len(t, request.chats, 2)
	require.Equal(t, constant.TDLibPeerID(-1000000000003), request.chats[0].id)
	require.Equal(t, constant.TDLibPeerID(-2), request.chats[1].id)
}
