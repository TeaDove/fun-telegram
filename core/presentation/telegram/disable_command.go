package telegram

import (
	"context"
	"strconv"
	"strings"

	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/core/repository/mongo_repository"
	"github.com/teadove/goteleout/core/repository/redis_repository"
)

func (r *Presentation) isEnabled(ctx context.Context, chatId int64) (bool, error) {
	_, err := r.redisRepository.Load(ctx, strconv.Itoa(int(chatId)))
	if err != nil {
		if errors.Is(err, redis_repository.ErrKeyNotFound) {
			return true, nil
		} else {
			return false, errors.Wrap(err, "failed to load from redis repository")
		}
	}

	return false, nil
}

func (r *Presentation) isBanned(ctx context.Context, username string) (bool, error) {
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

func (r *Presentation) disableCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	chatId := strconv.Itoa(int(update.EffectiveChat().GetID()))

	_, err := r.redisRepository.Load(ctx, chatId)
	if err != nil {
		if errors.Is(err, redis_repository.ErrKeyNotFound) {
			err = r.redisRepository.Save(ctx, chatId, []byte("1"))
			if err != nil {
				return errors.WithStack(err)
			}

			if !input.Silent {
				_, err = ctx.Reply(update, "Bot disabled in this chat", nil)
				if err != nil {
					return errors.WithStack(err)
				}
			}
		}

		return errors.WithStack(err)
	}

	err = r.redisRepository.Delete(ctx, chatId)
	if err != nil {
		return errors.WithStack(err)
	}

	if !input.Silent {
		_, err = ctx.Reply(update, "Bot enabled in this chat", nil)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}
