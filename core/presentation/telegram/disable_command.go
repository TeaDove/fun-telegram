package telegram

import (
	"context"
	"strings"

	"github.com/teadove/fun_telegram/core/repository/db_repository"

	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/redis_repository"
)

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
	if r.checkFromOwner(ctx, update) {
		return true, nil
	}

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
				zerolog.Ctx(ctx).
					Error().
					Stack().
					Err(err).
					Str("status", "failed.to.update.members").
					Send()
			}
		}()

		return false, errors.New("user not found in members")
	}

	return userMember.Status == db_repository.Admin ||
		userMember.Status == db_repository.Creator, nil
}

func (r *Presentation) checkFromOwner(ctx *ext.Context, update *ext.Update) (ok bool) {
	return update.EffectiveUser().GetID() == ctx.Self.ID
}
