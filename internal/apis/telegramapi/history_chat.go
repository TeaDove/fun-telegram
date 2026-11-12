package telegramapi

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/telegram/query"
	"github.com/gotd/td/telegram/query/dialogs"
	"github.com/gotd/td/tg"
)

const historyPeerIDFlag = "peer-id"

type historyChat struct {
	peer tg.InputPeerClass
	id   constant.TDLibPeerID
	name string
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

type historyChatSelector struct {
	id   constant.TDLibPeerID
	name string
}

func parseHistoryChatSelector(text string, options map[string]string, saved bool) (historyChatSelector, error) {
	value, hasID := options[historyPeerIDFlag]
	if !saved {
		if text != "" || hasID {
			return historyChatSelector{}, errors.New("specify a chat only when running the command from Saved Messages")
		}

		return historyChatSelector{}, nil
	}

	if !hasID && text != "" {
		return historyChatSelector{name: text}, nil
	}

	if !hasID || text != "" {
		return historyChatSelector{}, errors.New("specify one chat title or --peer-id=<peer_id>")
	}

	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return historyChatSelector{}, errors.Wrap(err, "parse peer ID")
	}

	if id == 0 {
		return historyChatSelector{}, errors.New("peer ID must be nonzero")
	}

	return historyChatSelector{id: constant.TDLibPeerID(id)}, nil
}

func (s historyChatSelector) matches(chat historyChat) bool {
	if s.id != 0 {
		return s.id == chat.id
	}

	return strings.EqualFold(s.name, strings.TrimSpace(chat.name))
}

func selectHistoryChat(matches map[constant.TDLibPeerID]historyChat) (historyChat, error) {
	if len(matches) == 0 {
		return historyChat{}, errors.New("chat not found in your Telegram dialogs")
	}

	if len(matches) == 1 {
		for _, chat := range matches {
			return chat, nil
		}
	}

	ids := make([]constant.TDLibPeerID, 0, len(matches))
	for id := range matches {
		ids = append(ids, id)
	}

	slices.Sort(ids)

	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, strconv.FormatInt(int64(id), 10))
	}

	return historyChat{}, errors.Newf("multiple chats match the title. Use --peer-id with one of: %s",
		strings.Join(values, ", "))
}

func (r *Presentation) resolveHistoryChat(c *Context) (historyChat, error) {
	selector, err := parseHistoryChatSelector(c.Text, c.Ops, c.inSavedMessages())
	if err != nil {
		return historyChat{}, err
	}

	if c.inSavedMessages() {
		return r.findHistoryChat(c.extCtx, selector, constant.TDLibPeerID(c.extCtx.Self.ID))
	}

	p := c.update.EffectiveMessage.PeerID

	id := telegramPeerID(p)
	if id == 0 {
		return historyChat{}, errors.New("unable to identify the current chat")
	}

	input := c.update.EffectiveChat().GetInputPeer()
	if _, empty := input.(*tg.InputPeerEmpty); empty || input == nil {
		return historyChat{}, errors.New("current chat is unavailable")
	}

	name := GetChatName(c.update.EffectiveChat())
	if c.update.Entities != nil {
		if title := historyPeerName(p, peer.EntitiesFromUpdate(*c.update.Entities)); title != "" {
			name = title
		}
	}

	return historyChat{peer: input, id: id, name: name}, nil
}

func historyChatFromDialog(elem dialogs.Elem) historyChat {
	dialog, ok := elem.Dialog.(*tg.Dialog)
	if !ok {
		return historyChat{}
	}

	if _, empty := elem.Peer.(*tg.InputPeerEmpty); empty || elem.Peer == nil {
		return historyChat{}
	}

	return historyChat{
		peer: elem.Peer,
		id:   telegramPeerID(dialog.Peer),
		name: historyPeerName(dialog.Peer, elem.Entities),
	}
}

func (r *Presentation) findHistoryChat(
	ctx context.Context,
	selector historyChatSelector,
	selfID constant.TDLibPeerID,
) (historyChat, error) {
	const archiveFolder = 1

	matches := make(map[constant.TDLibPeerID]historyChat)

	for _, folder := range []int{0, archiveFolder} {
		iter := query.GetDialogs(r.telegramAPI).FolderID(folder).BatchSize(historyBatchSize).Iter()
		for iter.Next(ctx) {
			chat := historyChatFromDialog(iter.Value())
			if chat.id == 0 || chat.id == selfID || !selector.matches(chat) {
				continue
			}

			if selector.id != 0 {
				return chat, nil
			}

			matches[chat.id] = chat
		}

		err := iter.Err()
		if err != nil {
			return historyChat{}, errors.Wrap(err, "read Telegram dialogs")
		}
	}

	return selectHistoryChat(matches)
}

func (c historyChat) label() string {
	if c.name == "" {
		return fmt.Sprintf("peer_id: %d", c.id)
	}

	return fmt.Sprintf("%s\npeer_id: %d", c.name, c.id)
}
