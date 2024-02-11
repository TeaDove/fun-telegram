package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/repository/redis_repository"
	"github.com/teadove/goteleout/internal/service/resource"
	"strings"
)

func compileBanPath(username string) string {
	return fmt.Sprintf("ban::%s", username)
}

func (r *Presentation) banCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	usernameToBanLower := strings.ToLower(input.Text)
	if usernameToBanLower == "" {
		err := r.replyIfNotSilentLocalized(ctx, update, input, resource.ErrUsernameRequired)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	selfUsername := strings.ToLower(ctx.Self.Username)
	if usernameToBanLower == selfUsername {
		if strings.ToLower(update.EffectiveUser().Username) == selfUsername {
			_, err := ctx.Reply(update, "?????????", nil)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		_, err := ctx.Reply(update, r.resourceService.Localize(ctx, resource.ErrNiceTry, input.Locale), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		usernameToBanLower = update.EffectiveUser().Username

		err = r.redisRepository.Save(compileBanPath(usernameToBanLower), []byte{})
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = ctx.Reply(update, r.resourceService.Localizef(ctx, resource.CommandBanUserBanned, input.Locale, usernameToBanLower), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "user.banned").Str("username", usernameToBanLower).Send()

		return nil
	}

	if update.EffectiveUser().GetID() != ctx.Self.ID {
		_, err := ctx.Reply(update, r.resourceService.Localize(ctx, resource.ErrInsufficientPrivilegesOwner, input.Locale), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	username := compileBanPath(usernameToBanLower)

	_, err := r.redisRepository.Load(username)
	if err != nil {
		if errors.Is(err, redis_repository.ErrKeyNotFound) {
			err = r.redisRepository.Save(username, []byte{})
			if err != nil {
				return errors.WithStack(err)
			}

			_, err = ctx.Reply(update, r.resourceService.Localizef(ctx, resource.CommandBanUserBanned, input.Locale, usernameToBanLower), nil)
			if err != nil {
				return errors.WithStack(err)
			}

			zerolog.Ctx(ctx).Info().Str("status", "user.banned").Str("username", usernameToBanLower).Send()
		} else {
			return errors.WithStack(err)
		}
	} else {
		err = r.redisRepository.Delete(username)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = ctx.Reply(update, r.resourceService.Localizef(ctx, resource.CommandBanUserUnbanned, input.Locale, usernameToBanLower), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "user.unbanned").Str("username", usernameToBanLower).Send()
	}

	return nil
}
