package telegramapi

import (
	"cmp"
	"slices"
	"strings"

	"github.com/cockroachdb/errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gotd/td/constant"
)

func historyChatNames(c *Context) ([]string, error) {
	names := make([]string, 0, len(c.Ops[historyNameFlag])+1)
	if c.Text != "" {
		names = append(names, c.Text)
	}

	for _, name := range c.Ops[historyNameFlag] {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("specify a chat name after -n")
		}

		names = append(names, name)
	}

	return names, nil
}

func selectHistoryChats(names []string, catalog map[constant.TDLibPeerID]historyChat) (historyChats, error) {
	chats := make(historyChats, 0)
	seen := mapset.NewThreadUnsafeSet[constant.TDLibPeerID]()

	for _, name := range names {
		matches := make(historyChats, 0)

		for _, chat := range catalog {
			if strings.EqualFold(strings.TrimSpace(chat.name), strings.TrimSpace(name)) {
				matches = append(matches, chat)
			}
		}

		if len(matches) == 0 {
			return nil, errors.Newf("chat %q not found in your Telegram dialogs", name)
		}

		slices.SortFunc(matches, func(a, b historyChat) int {
			return cmp.Compare(a.id, b.id)
		})

		for _, chat := range matches {
			if seen.Add(chat.id) {
				chats = append(chats, chat)
			}
		}
	}

	return chats, nil
}

func (r *Presentation) resolveHistoryChats(c *Context) (historyChats, error) {
	names, err := historyChatNames(c)
	if err != nil {
		return nil, err
	}

	var current historyChat

	if len(names) == 0 {
		if c.inSavedMessages() {
			return nil, errors.New("specify a chat name with -n")
		}

		current, err = r.resolveHistoryChat(c)
		if err != nil {
			return nil, err
		}

		names = []string{current.name}
	}

	catalog, err := r.historyDialogs(c.extCtx, c.extCtx.Self.ID)
	if err != nil {
		return nil, err
	}

	if current.id != 0 {
		catalog[current.id] = current
	}

	return selectHistoryChats(names, catalog)
}
