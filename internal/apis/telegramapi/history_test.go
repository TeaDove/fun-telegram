package telegramapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
	"github.com/stretchr/testify/require"
)

func TestDumpHistoryCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		saved   bool
	}{
		{name: "archived chat title", command: `!dump_history -q "Тестовый чат"`, saved: true},
		{name: "archived chat ID", command: "!dump_history --peer-id=-2", saved: true},
		{name: "quiet command in chat", command: "!dump_history -q"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			testDumpHistoryCommand(t, test.command, test.saved)
		})
	}
}

func testDumpHistoryCommand(t *testing.T, command string, saved bool) {
	t.Helper()

	const (
		commandID  = 7
		messageID  = 6
		progressID = 8
	)

	mock := tgmock.NewRequire(t)
	api := tg.NewClient(mock)
	self := &tg.User{ID: 1}
	source := &tg.Chat{ID: 2, Title: "Тестовый чат", Photo: &tg.ChatPhotoEmpty{}}
	sourcePeer := &tg.PeerChat{ChatID: source.ID}

	origin := tg.PeerClass(sourcePeer)
	offset := commandID

	if saved {
		origin = &tg.PeerUser{UserID: self.ID}
		offset = 0

		mock.ExpectCall(&tg.MessagesGetDialogsRequest{
			Limit: historyBatchSize, OffsetPeer: &tg.InputPeerEmpty{},
		}).ThenResult(&tg.MessagesDialogs{})

		mock.ExpectCall(&tg.MessagesGetDialogsRequest{
			Limit: historyBatchSize, FolderID: 1, OffsetPeer: &tg.InputPeerEmpty{},
		}).ThenResult(&tg.MessagesDialogs{
			Dialogs: []tg.DialogClass{&tg.Dialog{Peer: sourcePeer}},
			Chats:   []tg.ChatClass{source},
		})

		if !strings.Contains(command, "--peer-id") {
			mock.ExpectCall(&tg.MessagesGetDialogsRequest{
				Limit: historyBatchSize, FolderID: 1, OffsetPeer: &tg.InputPeerEmpty{},
			}).ThenResult(&tg.MessagesDialogs{})
		}
	} else {
		mock.ExpectCall(&tg.MessagesDeleteMessagesRequest{Revoke: true, ID: []int{commandID}}).
			ThenResult(&tg.MessagesAffectedMessages{})
	}

	mock.ExpectFunc(func(body bin.Encoder) {
		require.IsType(t, &tg.MessagesSendMessageRequest{}, body)
		request := body.(*tg.MessagesSendMessageRequest)
		require.IsType(t, &tg.InputPeerSelf{}, request.Peer)
		require.Contains(t, request.Message, "peer_id: -2")
	}).ThenResult(&tg.UpdateShortSentMessage{ID: progressID})

	author := &tg.User{ID: 3, FirstName: "Автор", Username: "test_author"}
	date := time.Now().UTC().Truncate(time.Second).Add(-time.Hour)
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer: &tg.InputPeerChat{ChatID: source.ID}, OffsetID: offset, Limit: historyBatchSize,
	}).ThenResult(&tg.MessagesMessages{
		Chats: []tg.ChatClass{source},
		Users: []tg.UserClass{author},
		Messages: []tg.MessageClass{&tg.Message{
			ID: messageID, PeerID: sourcePeer, Date: int(date.Unix()), Message: "Строка\n<тест>",
			FromID:  &tg.PeerUser{UserID: author.ID},
			ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: messageID - 1},
		}},
	})
	mock.ExpectCall(&tg.MessagesGetHistoryRequest{
		Peer: &tg.InputPeerChat{ChatID: source.ID}, OffsetID: messageID,
		OffsetDate: int(date.Unix()), Limit: historyBatchSize,
	}).ThenResult(&tg.MessagesMessages{})
	mock.ExpectFunc(func(body bin.Encoder) {
		require.IsType(t, &tg.MessagesEditMessageRequest{}, body)
		request := body.(*tg.MessagesEditMessageRequest)
		require.IsType(t, &tg.InputPeerSelf{}, request.Peer)
		require.Equal(t, progressID, request.ID)
		require.Contains(t, request.Message, "Messages read: 1")
	}).ThenResult(&tg.Updates{})

	var archive []byte

	mock.ExpectFunc(func(body bin.Encoder) {
		require.IsType(t, &tg.UploadSaveFilePartRequest{}, body)
		archive = bytes.Clone(body.(*tg.UploadSaveFilePartRequest).Bytes)
	}).ThenTrue()

	mock.ExpectFunc(func(body bin.Encoder) {
		require.IsType(t, &tg.MessagesSendMediaRequest{}, body)
		request := body.(*tg.MessagesSendMediaRequest)
		require.IsType(t, &tg.InputPeerSelf{}, request.Peer)
		require.Contains(t, request.Message, "peer_id: -2")
		require.IsType(t, &tg.InputMediaUploadedDocument{}, request.Media)
		document := request.Media.(*tg.InputMediaUploadedDocument)
		require.Equal(t, "application/gzip", document.MimeType)
		require.Contains(t, document.Attributes, &tg.DocumentAttributeFilename{FileName: historyDumpFilename})
	}).ThenResult(&tg.Updates{})

	r := Presentation{telegramAPI: api}
	r.router = map[string]messageProcessor{
		"dump_history": {executor: r.dumpHistoryCommand, flags: historyFlags()},
	}
	entities := &tg.Entities{Chats: map[int64]*tg.Chat{source.ID: source}}
	ctx := ext.NewContext(t.Context(), api, nil, self, message.NewSender(api), entities, false)
	update := &ext.Update{
		EffectiveMessage: types.ConstructMessage(&tg.Message{
			ID: commandID, Out: true, PeerID: origin, Message: command,
		}),
		Entities: entities,
	}

	err := r.route(ctx, update)
	require.NoError(t, err)
	requireHistoryArchive(t, archive, date, messageID)
}

func requireHistoryArchive(t *testing.T, archive []byte, date time.Time, messageID int) {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(archive))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reader.Close()) })

	decoder := json.NewDecoder(reader)

	var header historyDumpHeader
	require.NoError(t, decoder.Decode(&header))
	require.Equal(t, historyDumpHeader{Type: "chat", Version: 1, PeerID: "-2", Title: "Тестовый чат"}, header)

	var author historyDumpAuthor
	require.NoError(t, decoder.Decode(&author))
	require.Equal(t, historyDumpAuthor{Type: "author", ID: "3", Name: "Автор", Username: "test_author"}, author)

	var message historyDumpMessage
	require.NoError(t, decoder.Decode(&message))
	require.Equal(t, historyDumpMessage{
		Type: "message", ID: messageID, Date: date.Unix(), AuthorID: "3", Text: "Строка\n<тест>", ReplyTo: messageID - 1,
	}, message)
	require.ErrorIs(t, decoder.Decode(&message), io.EOF)
}

func TestHistoryChatSelector(t *testing.T) {
	t.Parallel()

	id := telegramPeerID(&tg.PeerChannel{ChannelID: 2})
	args := getOpt("!stats --peer-id=-1000000000002 -d=7", historyFlags()...)
	selector, err := parseHistoryChatSelector(args.Text, args.Ops, true)
	require.NoError(t, err)
	require.True(t, selector.matches(historyChat{id: id, name: "Renamed"}))
	require.False(t, selector.matches(historyChat{id: -2, name: "Renamed"}))

	_, err = parseHistoryChatSelector("", nil, true)
	require.Error(t, err)

	_, err = parseHistoryChatSelector("Title", args.Ops, true)
	require.Error(t, err)

	_, err = parseHistoryChatSelector("", args.Ops, false)
	require.Error(t, err)

	_, err = selectHistoryChat(map[constant.TDLibPeerID]historyChat{
		id: {id: id, name: "Title"}, -2: {id: -2, name: "Title"},
	})
	require.ErrorContains(t, err, "-1000000000002, -2")
}

func TestWalkHistoryLimits(t *testing.T) {
	t.Parallel()

	const cutoffDate = 20

	tests := []struct {
		name  string
		count int
		stop  string
		ids   []int
	}{
		{name: "count limit", count: 1, stop: "message limit", ids: []int{4}},
		{name: "date limit includes boundary", count: 5, stop: "date limit", ids: []int{4, 2}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			iter := &testHistoryIterator{records: []tg.NotEmptyMessage{
				&tg.Message{ID: 4, Date: cutoffDate + 1},
				&tg.MessageService{ID: 3, Date: cutoffDate},
				&tg.Message{ID: 2, Date: cutoffDate},
				&tg.Message{ID: 1, Date: cutoffDate - 1},
			}}

			ids := make([]int, 0, test.count)
			summary, err := walkHistory(t.Context(), iter,
				historyOptions{count: test.count, cutoff: time.Unix(cutoffDate, 0)}, 1,
				func(record historyMessage) error {
					ids = append(ids, record.id)
					return nil
				}, func(context.Context, string) {})
			require.NoError(t, err)
			require.Equal(t, test.ids, ids)
			require.Equal(t, len(test.ids), summary.count)
			require.Equal(t, test.stop, summary.stop)
		})
	}
}

type testHistoryIterator struct {
	records []tg.NotEmptyMessage
	index   int
}

func (i *testHistoryIterator) Next(context.Context) bool {
	i.index++
	return i.index <= len(i.records)
}

func (i *testHistoryIterator) Value() messages.Elem {
	return messages.Elem{Msg: i.records[i.index-1]}
}

func (i *testHistoryIterator) Err() error {
	return nil
}
