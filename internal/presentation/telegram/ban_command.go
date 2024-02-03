package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	tgUtils "github.com/teadove/goteleout/internal/presentation/telegram/utils"
	"github.com/teadove/goteleout/internal/service/storage"
	"strings"
)

func compileBanKey(username string) string {
	return fmt.Sprintf("ban::%s", username)
}

func (r *Presentation) banCommandHandler(ctx *ext.Context, update *ext.Update, input *tgUtils.Input) error {
	usernameToBanLower := strings.ToLower(input.Text)
	if usernameToBanLower == "" {
		_, err := ctx.Reply(update, "Err: no username found", nil)
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

		_, err := ctx.Reply(update, "Nice try", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		usernameToBanLower = update.EffectiveUser().Username

		err = r.storage.Save(compileBanKey(usernameToBanLower), []byte{})
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = ctx.Reply(update, fmt.Sprintf("%s was banned", usernameToBanLower), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "user.banned").Str("username", usernameToBanLower).Send()

		return nil
	}

	if update.EffectiveUser().GetID() != ctx.Self.ID {
		_, err := ctx.Reply(update, "Err: user can be banned only by owner of bot", nil)
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	username := compileBanKey(usernameToBanLower)

	_, err := r.storage.Load(username)
	if err != nil {
		if errors.Is(err, storage.ErrKeyNotFound) {
			err = r.storage.Save(username, []byte{})
			if err != nil {
				return errors.WithStack(err)
			}

			_, err = ctx.Reply(update, fmt.Sprintf("%s was banned", usernameToBanLower), nil)
			if err != nil {
				return errors.WithStack(err)
			}

			zerolog.Ctx(ctx).Info().Str("status", "user.banned").Str("username", usernameToBanLower).Send()
		} else {
			return errors.WithStack(err)
		}
	} else {
		err = r.storage.Delete(username)
		if err != nil {
			return errors.WithStack(err)
		}

		_, err = ctx.Reply(update, fmt.Sprintf("%s was unbanned", usernameToBanLower), nil)
		if err != nil {
			return errors.WithStack(err)
		}

		zerolog.Ctx(ctx).Info().Str("status", "user.unbanned").Str("username", usernameToBanLower).Send()
	}

	return nil
}
