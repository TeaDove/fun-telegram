package telegramapi

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
	"github.com/stretchr/testify/require"
)

func TestDumpMultipleChats(t *testing.T) {
	t.Parallel()

	chats := []tg.ChatClass{
		&tg.Chat{ID: 2, Title: "First test chat", Photo: &tg.ChatPhotoEmpty{}},
		&tg.Chat{ID: 3, Title: "Second test chat", Photo: &tg.ChatPhotoEmpty{}},
	}
	self := &tg.User{ID: 1, FirstName: "Author"}
	date := time.Now().UTC().Add(-time.Hour).Unix()

	var archive []byte

	sentFiles := 0
	api := tg.NewClient(tgmock.Invoker(func(request bin.Encoder) (bin.Encoder, error) {
		switch request := request.(type) {
		case *tg.MessagesGetDialogsRequest:
			if request.FolderID != 0 {
				return &tg.MessagesDialogs{}, nil
			}

			return &tg.MessagesDialogs{Chats: chats, Dialogs: []tg.DialogClass{
				&tg.Dialog{Peer: &tg.PeerChat{ChatID: 2}}, &tg.Dialog{Peer: &tg.PeerChat{ChatID: 3}},
			}}, nil
		case *tg.MessagesSendMessageRequest:
			require.IsType(t, &tg.InputPeerSelf{}, request.Peer)
			return &tg.UpdateShortSentMessage{ID: 10}, nil
		case *tg.MessagesGetHistoryRequest:
			if request.OffsetID != 0 {
				return &tg.MessagesMessages{}, nil
			}

			id := request.Peer.(*tg.InputPeerChat).ChatID

			return &tg.MessagesMessages{Chats: chats, Users: []tg.UserClass{self}, Messages: []tg.MessageClass{
				&tg.Message{ID: 5, Date: int(date - id), PeerID: &tg.PeerChat{ChatID: id},
					FromID: &tg.PeerUser{UserID: self.ID}, Message: "Test message"},
			}}, nil
		case *tg.MessagesEditMessageRequest:
			require.Contains(t, request.Message, "msgs=1, range=0s")
			return &tg.Updates{}, nil
		case *tg.UploadSaveFilePartRequest:
			archive = append(archive, request.Bytes...)
			return &tg.BoolTrue{}, nil
		case *tg.MessagesSendMediaRequest:
			require.IsType(t, &tg.InputPeerSelf{}, request.Peer)
			require.Contains(t, request.Message, "First test chat (-2), msgs=1, range=0s")
			require.Contains(t, request.Message, "Second test chat (-3), msgs=1, range=0s")

			sentFiles++

			return &tg.Updates{}, nil
		default:
			t.Fatalf("unexpected Telegram request %T", request)
			return nil, nil
		}
	}))
	presentation := Presentation{telegramAPI: api}
	presentation.router = map[string]messageProcessor{
		"dump-history": {executor: presentation.dumpHistoryCommand, flags: historyFlags()},
	}
	ctx := ext.NewContext(t.Context(), api, nil, self, message.NewSender(api), nil, false)
	update := &ext.Update{EffectiveMessage: types.ConstructMessage(&tg.Message{
		ID: 8, PeerID: &tg.PeerUser{UserID: self.ID},
		Message: `!dump-history -n "First test chat" -n "Second test chat" -n "First test chat" -c=2`,
	})}
	err := presentation.route(ctx, update)
	require.NoError(t, err)
	require.Equal(t, 1, sentFiles)

	assertBulkArchive(t, archive)
}

func assertBulkArchive(t *testing.T, archive []byte) {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(archive))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reader.Close()) })

	decoder := json.NewDecoder(reader)
	peers := make([]int64, 0, 2)

	for range 2 {
		var raw json.RawMessage
		require.NoError(t, decoder.Decode(&raw))
		require.NotContains(t, string(raw), "version")

		var header tgdtos.Record[tgdtos.HistoryHeader]
		require.NoError(t, json.Unmarshal(raw, &header))
		peers = append(peers, header.Payload.PeerID)
	}

	require.Equal(t, []int64{-2, -3}, peers)

	var author tgdtos.Record[tgdtos.Author]
	require.NoError(t, decoder.Decode(&author))
	require.Equal(t, tgdtos.TGTypeAuthor, author.Type)

	for _, peerID := range peers {
		var msg tgdtos.Record[tgdtos.Message]
		require.NoError(t, decoder.Decode(&msg))
		require.Equal(t, tgdtos.TGTypeMessage, msg.Type)
		require.Equal(t, peerID, msg.Payload.PeerID)
		require.Equal(t, 5, msg.Payload.ID)
	}

	require.ErrorIs(t, decoder.Decode(&author), io.EOF)
}

func TestDumpTopicIDsBelongToPeer(t *testing.T) {
	t.Parallel()

	chats := historyChats{
		{id: -1000000000002, peer: &tg.InputPeerChannel{ChannelID: 2}},
		{id: -1000000000003, peer: &tg.InputPeerChannel{ChannelID: 3}},
	}
	mock := tgmock.NewRequire(t)

	for _, chat := range chats {
		channel := chat.peer.(*tg.InputPeerChannel)
		mock.ExpectCall(&tg.MessagesGetForumTopicsByIDRequest{Peer: chat.peer, Topics: []int{5}}).
			ThenResult(&tg.MessagesForumTopics{Topics: []tg.ForumTopicClass{
				&tg.ForumTopic{
					ID: 5, Title: "Test topic", Peer: &tg.PeerChannel{ChannelID: channel.ChannelID},
					FromID: &tg.PeerUser{UserID: 1},
				},
			}})
	}

	var output bytes.Buffer

	dump, err := newHistoryDump(&output, chats)
	require.NoError(t, err)

	for _, chat := range chats {
		err = dump.append(tgdtos.Message{PeerID: int64(chat.id), ID: 6, TopicID: 5}, tgdtos.Author{})
		require.NoError(t, err)
	}

	err = dump.appendTopics(t.Context(), tg.NewClient(mock), chats)
	require.NoError(t, err)

	decoder := json.NewDecoder(&output)

	for range 4 {
		var record json.RawMessage
		require.NoError(t, decoder.Decode(&record))
	}

	for _, chat := range chats {
		var topic tgdtos.Record[tgdtos.Topic]
		require.NoError(t, decoder.Decode(&topic))
		require.Equal(t, tgdtos.TGTypeTopic, topic.Type)
		require.Equal(t, int64(chat.id), topic.Payload.PeerID)
		require.Equal(t, 5, topic.Payload.ID)
	}
}
