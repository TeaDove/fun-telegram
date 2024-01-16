package telegram

import (
	"context"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/peers/members"
	"github.com/pkg/errors"
)

func (r *Presentation) getMembers(
	ctx context.Context,
	effectiveChat types.EffectiveChat,
) (map[int64]members.Member, error) {
	chatMembersSlice := make(map[int64]members.Member, 20)

	compileSlice := func(p members.Member) error {
		chatMembersSlice[p.User().ID()] = p

		return nil
	}

	switch t := effectiveChat.(type) {
	case *types.Chat:
		chat := r.telegramManager.Chat(t.Raw())
		chatMembers := members.Chat(chat)

		err := chatMembers.ForEach(ctx, compileSlice)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case *types.Channel:
		chat := r.telegramManager.Channel(t.Raw())
		chatMembers := members.Channel(chat)

		err := chatMembers.ForEach(ctx, compileSlice)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	default:
		return nil, errors.New("is not chat or channel")
	}

	return chatMembersSlice, nil
}
