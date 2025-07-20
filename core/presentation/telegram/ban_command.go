package telegram

import (
	"fmt"
	"strings"

	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/redis_repository"
)

func compileBanPath(username string) string {
	return fmt.Sprintf("ban::%s", username)
}

// banCommandHandler
// nolint: gocyclo
func (r *Presentation) banCommandHandler(ctx *ext.Context, update *ext.Update, input *input) error {
	usernameToBanLower := strings.ToLower(input.Text)
	if usernameToBanLower == "" {
		err := r.replyIfNotSilent(ctx, update, input, "Err: Username required")
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

		_, err := ctx.Reply(update, "Err: nice try", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		usernameToBanLower = update.EffectiveUser().Username

		err = r.redisRepository.Save(ctx, compileBanPath(usernameToBanLower), []byte{})
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = ctx.Reply(update, fmt.Sprintf("%s was banned", usernameToBanLower), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).
			Info().
			Str("username", usernameToBanLower).
			Msg("user.banned")

		return nil
	}

	if update.EffectiveUser().GetID() != ctx.Self.ID {
		_, err := ctx.Reply(
			update,
			"Err: insufficient privilege: owner rights required",
			nil,
		)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	username := compileBanPath(usernameToBanLower)

	_, err := r.redisRepository.Load(ctx, username)
	if err != nil {
		if errors.Is(err, redis_repository.ErrKeyNotFound) {
			err = r.redisRepository.Save(ctx, username, []byte{})
			if err != nil {
				return errors.WithStack(err)
			}

			_, err = ctx.Reply(
				update,
				fmt.Sprintf("%s was banned", usernameToBanLower),
				nil,
			)
			if err != nil {
				return errors.WithStack(err)
			}

			zerolog.Ctx(ctx).
				Info().
				Str("username", usernameToBanLower).
				Msg("user.banned")
			return nil

		}

		return errors.WithStack(err)
	}

	err = r.redisRepository.Delete(ctx, username)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = ctx.Reply(update, fmt.Sprintf("%s was unbanned", usernameToBanLower), nil)
	if err != nil {
		return errors.WithStack(err)
	}

	zerolog.Ctx(ctx).Info().Str("username", usernameToBanLower).Msg("user.unbanned")

	return nil
}
