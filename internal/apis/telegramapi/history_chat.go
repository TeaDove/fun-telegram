package telegramapi

import (
	"context"
	"strings"

	"github.com/TeaDove/fun-telegram/internal/dtos/tgdtos"
	"github.com/cockroachdb/errors"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

const historyNameFlag = "name"

type historyChat struct {
	peer       tg.InputPeerClass
	id         constant.TDLibPeerID
	name       string
	forum      bool
	migratedTo constant.TDLibPeerID
}

func telegramPeerID(p tg.PeerClass) constant.TDLibPeerID {
	var id constant.TDLibPeerID

	switch v := p.(type) {
	case *tg.PeerUser:
		id.User(v.UserID)
	case *tg.PeerChat:
		id.Chat(v.ChatID)
	case *tg.PeerChannel:
		id.Channel(v.ChannelID)
	}

	return id
}

func userDisplayName(user *tg.User) string {
	name := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if name == "" {
		return user.Username
	}

	return name
}

func historyPeerName(p tg.PeerClass, entities peer.Entities) string {
	switch v := p.(type) {
	case *tg.PeerUser:
		if user, ok := entities.User(v.UserID); ok {
			return userDisplayName(user)
		}
	case *tg.PeerChat:
		if chat, ok := entities.Chat(v.ChatID); ok {
			return chat.Title
		}
	case *tg.PeerChannel:
		if channel, ok := entities.Channel(v.ChannelID); ok {
			return channel.Title
		}
	}

	return ""
}

func (r *Presentation) resolveHistoryChat(c *Context) (historyChat, error) {
	effectiveChat := c.update.EffectiveChat()

	var entities peer.Entities
	if c.update.Entities != nil {
		entities = peer.EntitiesFromUpdate(*c.update.Entities)
	}

	chat := historyChatFromPeer(c.update.EffectiveMessage.PeerID, effectiveChat.GetInputPeer(), entities)
	if chat.id == 0 {
		return historyChat{}, errors.New("unable to identify the current chat")
	}

	if _, empty := chat.peer.(*tg.InputPeerEmpty); empty || chat.peer == nil {
		return historyChat{}, errors.New("current chat is unavailable")
	}

	if chat.name == "" {
		chat.name = GetChatName(effectiveChat)
	}

	return chat, nil
}

func historyChatFromPeer(p tg.PeerClass, input tg.InputPeerClass, entities peer.Entities) historyChat {
	chat := historyChat{peer: input, id: telegramPeerID(p), name: historyPeerName(p, entities)}

	if groupPeer, ok := p.(*tg.PeerChat); ok {
		if group, found := entities.Chat(groupPeer.ChatID); found {
			if target, ok := group.MigratedTo.(*tg.InputChannel); ok {
				chat.migratedTo = telegramPeerID(&tg.PeerChannel{ChannelID: target.ChannelID})
			}
		}
	}

	if channelPeer, ok := p.(*tg.PeerChannel); ok {
		channel, found := entities.Channel(channelPeer.ChannelID)
		chat.forum = found && channel.Forum
	}

	return chat
}

func historyChatFromDialog(elem dialogs.Elem) historyChat {
	dialog, ok := elem.Dialog.(*tg.Dialog)
	if !ok {
		return historyChat{}
	}

	if _, empty := elem.Peer.(*tg.InputPeerEmpty); empty || elem.Peer == nil {
		return historyChat{}
	}

	return historyChatFromPeer(dialog.Peer, elem.Peer, elem.Entities)
}

func (r *Presentation) historyDialogs(ctx context.Context, selfID int64) (map[constant.TDLibPeerID]historyChat, error) {
	const archiveFolder = 1

	chats := make(map[constant.TDLibPeerID]historyChat)

	for _, folder := range []int{0, archiveFolder} {
		iter := query.GetDialogs(r.telegramAPI).FolderID(folder).BatchSize(historyBatchSize).Iter()
		for iter.Next(ctx) {
			chat := historyChatFromDialog(iter.Value())
			if chat.id != 0 && chat.id != constant.TDLibPeerID(selfID) {
				chats[chat.id] = chat
			}
		}

		err := iter.Err()
		if err != nil {
			return nil, errors.Wrap(err, "read Telegram dialogs")
		}
	}

	return chats, nil
}

type historyChats []historyChat

func (c historyChats) toDTOs() []tgdtos.Chat {
	byID := make(map[constant.TDLibPeerID]historyChat, len(c))
	for _, chat := range c {
		byID[chat.id] = chat
	}

	output := make([]tgdtos.Chat, 0, len(c))
	groups := make(map[constant.TDLibPeerID]int, len(c))

	for _, chat := range c {
		groupID := chat.id
		title := chat.name

		if target, found := byID[chat.migratedTo]; found {
			groupID = target.id
			title = target.name
		}

		index, found := groups[groupID]
		if !found {
			index = len(output)
			groups[groupID] = index

			output = append(output, tgdtos.Chat{Title: title})
		}

		output[index].PeerIDs = append(output[index].PeerIDs, int64(chat.id))
	}

	return output
}
