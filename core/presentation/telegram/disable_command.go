package telegram

import (
	"context"
	"fmt"
	"strings"

	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/mongo_repository"
	"github.com/teadove/fun_telegram/core/repository/redis_repository"
)

func getEnabledPath(chatId int64) string {
	return fmt.Sprintf("enabled::%d", chatId)
}

func (r *Presentation) isEnabled(ctx context.Context, chatId int64) (bool, error) {
	enabled, err := r.redisRepository.GetToggle(ctx, getEnabledPath(chatId))
	if err != nil {
		return false, errors.Wrap(err, "failed to load from redis repository")
	}

	return enabled, nil
}

func (r *Presentation) isBanned(ctx context.Context, username string) (bool, error) {
	// TODO rework to Toggle
	_, err := r.redisRepository.Load(ctx, compileBanPath(strings.ToLower(username)))
	if err != nil {
		if errors.Is(err, redis_repository.ErrKeyNotFound) {
			return false, nil
		} else {
			return false, errors.Wrap(err, "failed to load from redis repository")
		}
	}

	return true, nil
}

func (r *Presentation) checkFromAdmin(ctx *ext.Context, update *ext.Update) (ok bool, err error) {
	chatMembers, err := r.getOrUpdateMembers(ctx, update.EffectiveChat())
	if err != nil {
		if errors.Is(err, ErrNotChatOrChannel) {
			// Expects, that in private conversation everyone is admin
			return true, nil
		}
		return false, errors.WithStack(err)
	}

	userMember, ok := chatMembers.ToMap()[update.EffectiveUser().GetID()]
	if !ok {
		go func() {
			_, err = r.updateMembers(ctx, update.EffectiveChat())
			if err != nil {
				zerolog.Ctx(ctx).Error().Stack().Err(err).Str("status", "failed.to.update.members").Send()
			}
		}()
		return false, errors.New("user not found in members")
	}

	return userMember.Status == mongo_repository.Admin ||
		r.checkFromOwner(ctx, update) ||
		userMember.Status == mongo_repository.Creator, nil
}

func (r *Presentation) checkFromOwner(ctx *ext.Context, update *ext.Update) (ok bool) {
	return update.EffectiveUser().GetID() == ctx.Self.ID
}

func (r *Presentation) disableCommandHandler(ctx *ext.Context, update *ext.Update, input *input) error {
	wasEnabled, err := r.redisRepository.Toggle(ctx, getEnabledPath(update.EffectiveChat().GetID()))
	if err != nil {
		return errors.WithStack(err)
	}

	var text string
	if wasEnabled {
		text = "Bot disabled in this chat"
	} else {
		text = "Bot enabled in this chat"
	}

	err = r.replyIfNotSilent(ctx, update, input, text)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
