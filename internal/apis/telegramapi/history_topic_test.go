package telegramapi

import (
	"bytes"
	"strings"
	"testing"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgmock"
	"github.com/stretchr/testify/require"
)

func TestHistoryChatNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		command string
		want    string
	}{
		{command: `!stats "Test chat"`, want: `chat "Test chat" not found`},
		{command: "!stats --peer-id=-2", want: "chat with peer_id -2 not found"},
		{command: "!stats -peer-id=-2", want: `chat "-peer-id=-2" not found`},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			t.Parallel()

			args := getOpt(test.command, historyFlags()...)
			selector, err := parseHistoryChatSelector(args.Text, args.Ops, true)
			require.NoError(t, err)

			_, err = selector.selectChat(nil)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestHistoryMessageTopic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		reply  tg.MessageReplyHeaderClass
		forum  bool
		topic  int
		parent int
	}{
		{name: "general", forum: true, topic: 1},
		{
			name: "topic root", forum: true, topic: 5, parent: 5,
			reply: &tg.MessageReplyHeader{ForumTopic: true, ReplyToMsgID: 5},
		},
		{
			name: "reply in topic", forum: true, topic: 5, parent: 6,
			reply: &tg.MessageReplyHeader{ForumTopic: true, ReplyToTopID: 5, ReplyToMsgID: 6},
		},
		{
			name: "ordinary reply thread", parent: 6,
			reply: &tg.MessageReplyHeader{ReplyToTopID: 5, ReplyToMsgID: 6},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			msg := &tg.Message{PeerID: &tg.PeerChannel{ChannelID: 2}, ReplyTo: test.reply}
			record, _ := decodeHistoryMessage(messages.Elem{}, msg, 1, test.forum)
			require.Equal(t, test.topic, record.TopicID)
			require.Equal(t, test.parent, record.ReplyTo)
		})
	}
}

func TestHistoryDumpForumTopics(t *testing.T) {
	t.Parallel()

	const channelID = 2

	chatPeer := &tg.PeerChannel{ChannelID: channelID}
	input := &tg.InputPeerChannel{ChannelID: channelID}
	entities := peer.NewEntities(nil, nil, map[int64]*tg.Channel{channelID: {ID: channelID, Forum: true}})
	chat := historyChatFromPeer(chatPeer, input, entities)
	require.True(t, chat.forum)

	mock := tgmock.NewRequire(t)
	mock.ExpectCall(&tg.MessagesGetForumTopicsByIDRequest{Peer: input, Topics: []int{1, 5, 9}}).
		ThenResult(&tg.MessagesForumTopics{Topics: []tg.ForumTopicClass{
			&tg.ForumTopic{ID: 1, Title: "General", Peer: chatPeer, FromID: &tg.PeerUser{UserID: 1}},
			&tg.ForumTopic{ID: 5, Title: "Test topic", Peer: chatPeer, FromID: &tg.PeerUser{UserID: 1}},
			&tg.ForumTopicDeleted{ID: 9},
		}})

	var output bytes.Buffer

	dump, err := newHistoryDump(&output, chat)
	require.NoError(t, err)

	for index, id := range []int{5, 5, 1, 9} {
		err = dump.append(tgdtos.Message{ID: 10 + index, TopicID: id}, tgdtos.Author{})
		require.NoError(t, err)
	}

	err = dump.appendTopics(t.Context(), tg.NewClient(mock), input)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	require.Len(t, lines, 7)
	require.JSONEq(t, `{"type":"MESSAGE","payload":{"id":10,"date":0,"topic_id":5}}`, lines[1])
	require.JSONEq(t, `{"type":"MESSAGE","payload":{"id":12,"date":0,"topic_id":1}}`, lines[3])
	require.JSONEq(t, `{"type":"MESSAGE","payload":{"id":13,"date":0,"topic_id":9}}`, lines[4])
	require.JSONEq(t, `{"type":"TOPIC","payload":{"id":1,"title":"General"}}`, lines[5])
	require.JSONEq(t, `{"type":"TOPIC","payload":{"id":5,"title":"Test topic"}}`, lines[6])
}
