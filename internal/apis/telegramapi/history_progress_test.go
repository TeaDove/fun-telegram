package telegramapi

import (
	"testing"
	"time"
	"unicode/utf16"

	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
	"github.com/stretchr/testify/require"
)

func TestHistoryProgressFormatting(t *testing.T) {
	t.Parallel()

	const name = "Test 😀<&>"

	mock := tgmock.NewRequire(t)
	mock.ExpectFunc(func(body bin.Encoder) {
		request := body.(*tg.MessagesSendMessageRequest)
		require.Equal(t, "Loading history for "+name+" with 500000 amount and 730 day limit", request.Message)
		requireHistoryProgressEntities(t, request.Message, request.Entities, name, nil)
	}).ThenResult(&tg.UpdateShortSentMessage{ID: 7})
	mock.ExpectFunc(func(body bin.Encoder) {
		request := body.(*tg.MessagesEditMessageRequest)
		require.Equal(t, 7, request.ID)
		require.Equal(t,
			"Loading history for "+name+" with 500000 amount and 730 day limit\nRead 3 msgs\nStopped: end of history",
			request.Message,
		)
		requireHistoryProgressEntities(t, request.Message, request.Entities, name, []string{"3", "end of history"})
	}).ThenResult(&tg.Updates{})

	api := tg.NewClient(mock)
	c := getOpt("!stats -c=999999 -d=9999", historyFlags()...)
	c.StartedAt = time.Now().UTC()
	c.extCtx = ext.NewContext(t.Context(), api, nil, &tg.User{ID: 1}, message.NewSender(api), nil, false)
	c.update = &ext.Update{EffectiveMessage: types.ConstructMessage(&tg.Message{PeerID: &tg.PeerUser{UserID: 1}})}

	options, err := parseHistoryOptions(c.Ops, c.StartedAt)
	require.NoError(t, err)

	progress, err := c.startHistoryProgress([]string{name, "test 😀<&>"}, options)
	require.NoError(t, err)
	progress(t.Context(), 3, "end of history")
	progress(t.Context(), 3, "end of history")
}

func requireHistoryProgressEntities(
	t *testing.T,
	text string,
	entities []tg.MessageEntityClass,
	name string,
	codes []string,
) {
	t.Helper()

	encoded := utf16.Encode([]rune(text))
	actualName := ""

	var actualCodes []string

	for _, item := range entities {
		switch item := item.(type) {
		case *tg.MessageEntityItalic:
			actualName = string(utf16.Decode(encoded[item.Offset : item.Offset+item.Length]))
		case *tg.MessageEntityCode:
			actualCodes = append(actualCodes, string(utf16.Decode(encoded[item.Offset:item.Offset+item.Length])))
		}
	}

	require.Equal(t, name, actualName)
	require.ElementsMatch(t, codes, actualCodes)
}

func TestHistoryProgressReadCounter(t *testing.T) {
	t.Parallel()

	header := []styling.StyledTextOption{styling.Plain("Loading history")}
	text, entities, err := renderHistoryProgress(header, 500, "")
	require.NoError(t, err)
	require.Equal(t, "Loading history\nRead 500 msgs", text)
	requireHistoryProgressEntities(t, text, entities, "", []string{"500"})
}

func TestHistoryProgressRepliesInSourceChat(t *testing.T) {
	t.Parallel()

	mock := tgmock.NewRequire(t)
	mock.ExpectFunc(func(body bin.Encoder) {
		request := body.(*tg.MessagesSendMessageRequest)
		require.Equal(t, &tg.InputPeerChat{ChatID: 2}, request.Peer)
		require.Equal(t, &tg.InputReplyToMessage{ReplyToMsgID: 4}, request.ReplyTo)
		require.Equal(t, "Loading history for A with 10000 amount and 60 day limit", request.Message)
		requireHistoryProgressEntities(t, request.Message, request.Entities, "A", nil)
	}).ThenResult(&tg.UpdateShortSentMessage{ID: 7})

	api := tg.NewClient(mock)
	entities := &tg.Entities{Chats: map[int64]*tg.Chat{2: {ID: 2, Title: "A"}}}
	msg := &tg.Message{ID: 4, Out: true, PeerID: &tg.PeerChat{ChatID: 2}}
	c := Context{
		StartedAt: time.Now().UTC(),
		extCtx:    ext.NewContext(t.Context(), api, nil, &tg.User{ID: 1}, message.NewSender(api), entities, false),
		update: &ext.Update{
			EffectiveMessage: types.ConstructMessage(msg),
			UpdateClass:      &tg.UpdateNewMessage{Message: msg},
			Entities:         entities,
		},
	}
	options, err := parseHistoryOptions(nil, c.StartedAt)
	require.NoError(t, err)

	_, err = c.startHistoryProgress([]string{"A"}, options)
	require.NoError(t, err)
}
