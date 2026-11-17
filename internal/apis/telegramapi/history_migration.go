package telegramapi

import (
	"context"

	"github.com/cockroachdb/errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/telegram/message/peer"
	"github.com/gotd/td/tg"
)

func parseHistoryChatSelectors(text string, options map[string][]string, saved bool) ([]historyChatSelector, error) {
	values := options[historyPeerIDFlag]
	if !saved || len(values) == 0 {
		selector, err := parseHistoryChatSelector(text, options, saved)
		return []historyChatSelector{selector}, err
	}

	selectors := make([]historyChatSelector, 0, len(values))
	for _, value := range values {
		selector, err := parseHistoryChatSelector(text, map[string][]string{historyPeerIDFlag: {value}}, saved)
		if err != nil {
			return nil, err
		}

		selectors = append(selectors, selector)
	}

	return selectors, nil
}

func (r *Presentation) resolveHistoryChats(c *Context) (historyChats, error) {
	selectors, err := parseHistoryChatSelectors(c.Text, c.Ops, c.inSavedMessages())
	if err != nil {
		return nil, err
	}

	if !c.inSavedMessages() {
		chat, err := r.resolveHistoryChat(c)
		if err != nil {
			return nil, err
		}

		return r.expandHistoryChat(c.extCtx, chat, nil)
	}

	catalog, err := r.historyDialogs(c.extCtx, c.extCtx.Self.ID)
	if err != nil {
		return nil, err
	}

	chats := make(historyChats, 0, len(selectors))
	seen := mapset.NewThreadUnsafeSet[constant.TDLibPeerID]()

	for _, selector := range selectors {
		chat, err := r.findHistoryChat(c.extCtx, selector, catalog)
		if err != nil {
			return nil, err
		}

		if seen.Contains(chat.canonicalID()) {
			continue
		}

		expanded, err := r.expandHistoryChat(c.extCtx, chat, catalog)
		if err != nil {
			return nil, err
		}

		for _, source := range expanded {
			catalog[source.id] = source
			if seen.Add(source.id) {
				chats = append(chats, source)
			}
		}
	}

	return chats, nil
}

func (r *Presentation) findHistoryChat(
	ctx context.Context,
	selector historyChatSelector,
	catalog map[constant.TDLibPeerID]historyChat,
) (historyChat, error) {
	matches := make(map[constant.TDLibPeerID]historyChat)
	for id, chat := range catalog {
		if selector.matches(chat) {
			matches[id] = chat
		}
	}

	if len(matches) == 0 && selector.id.IsChat() {
		result, err := r.telegramAPI.MessagesGetChats(ctx, []int64{selector.id.ToPlain()})
		if err != nil {
			return historyChat{}, errors.Wrap(err, "resolve original group")
		}

		entities := peer.NewEntities(nil, result.MapChats().ChatToMap(), result.MapChats().ChannelToMap())

		input := &tg.InputPeerChat{ChatID: selector.id.ToPlain()}
		if _, found := entities.Chat(input.ChatID); found {
			matches[selector.id] = historyChatFromPeer(&tg.PeerChat{ChatID: input.ChatID}, input, entities)
		}
	}

	return selector.selectChat(matches)
}

func (r *Presentation) expandHistoryChat(
	ctx context.Context,
	chat historyChat,
	catalog map[constant.TDLibPeerID]historyChat,
) (historyChats, error) {
	channel, ok := chat.peer.(*tg.InputPeerChannel)
	if chat.migratedTo != nil {
		channel = &tg.InputPeerChannel{ChannelID: chat.migratedTo.ChannelID, AccessHash: chat.migratedTo.AccessHash}
		ok = true
	}

	if !ok {
		return historyChats{chat}, nil
	}

	result, err := r.telegramAPI.ChannelsGetFullChannel(ctx, &tg.InputChannel{
		ChannelID: channel.ChannelID, AccessHash: channel.AccessHash,
	})
	if err != nil {
		return nil, errors.Wrap(err, "read group migration")
	}

	full, ok := result.FullChat.(*tg.ChannelFull)
	if !ok {
		return nil, errors.New("unexpected supergroup information")
	}

	entities := peer.EntitiesFromResult(result)

	current := historyChatFromPeer(&tg.PeerChannel{ChannelID: channel.ChannelID}, channel, entities)
	if current.name == "" {
		current.name = chat.name
		current.forum = chat.forum
	}

	chats := historyChats{current}

	oldID := full.MigratedFromChatID
	if chat.migratedTo != nil {
		oldID = chat.id.ToPlain()
	}

	if oldID == 0 {
		return chats, nil
	}

	id := telegramPeerID(&tg.PeerChat{ChatID: oldID})

	old, found := catalog[id]
	if !found {
		old = historyChatFromPeer(&tg.PeerChat{ChatID: oldID}, &tg.InputPeerChat{ChatID: oldID}, entities)
	}

	if old.name == "" {
		old.name = current.name
	}

	old.migratedTo = &tg.InputChannel{ChannelID: channel.ChannelID, AccessHash: channel.AccessHash}

	return append(chats, old), nil
}
