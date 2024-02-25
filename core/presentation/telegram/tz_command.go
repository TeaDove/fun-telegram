package telegram

import (
	"context"
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/repository/redis_repository"
	"github.com/teadove/fun_telegram/core/service/resource"
	"strconv"
	"strings"
	"time"
)

// TODO move all chat settings to one place
func getTimezonePath(chatId int64) string {
	return fmt.Sprintf("tz::%d", chatId)
}

const tzOffset int8 = 12

func int8ToLoc(tz int8) *time.Location {
	if tz == 0 {
		return time.FixedZone("UTC", 0)
	}

	if tz > 0 {
		return time.FixedZone(fmt.Sprintf("UTC+%d", tz), int(tz)*60*60)
	}

	return time.FixedZone(fmt.Sprintf("UTC%d", tz), int(tz)*60*60)
}

func (r *Presentation) getTz(ctx context.Context, chatId int64) (int8, error) {
	path := getTimezonePath(chatId)

	tzBytes, err := r.redisRepository.Load(ctx, path)
	if err != nil {
		if errors.Is(err, redis_repository.ErrKeyNotFound) {
			println("key not found")
			return 0, nil
		}
		return 0, errors.Wrap(err, "failed to load from redis")
	}

	if len(tzBytes) != 1 {
		zerolog.Ctx(ctx).Warn().Str("status", "broken.tz.in.path.resetting.it").Send()
		err = r.redisRepository.Delete(ctx, path)
		if err != nil {
			return 0, errors.Wrap(err, "failed to delete tz")
		}

		return 0, nil
	}

	return int8(tzBytes[0]) - tzOffset, nil
}

func (r *Presentation) tzCommandHandler(ctx *ext.Context, update *ext.Update, input *input) error {
	tz, err := strconv.Atoi(strings.TrimSpace(input.Text))
	if err != nil {
		err := r.replyIfNotSilent(
			ctx,
			update,
			input,
			r.resourceService.Localizef(ctx, resource.ErrUnprocessableEntity, input.Locale, err),
		)
		if err != nil {
			return errors.Wrap(err, "failed to reply if not silent")
		}

		return nil
	}

	if !(-12 < tz && tz < 12) {
		err = r.replyIfNotSilent(
			ctx,
			update,
			input,
			r.resourceService.Localizef(ctx, resource.ErrUnprocessableEntity, input.Locale, errors.New("tz should be from -12 to 12")),
		)
		if err != nil {
			return errors.Wrap(err, "failed to reply if not silent")
		}

		return nil
	}
	tz8 := int8(tz)

	path := getTimezonePath(update.EffectiveChat().GetID())

	err = r.redisRepository.Save(ctx, path, []byte{byte(tz8 + tzOffset)})
	if err != nil {
		return errors.Wrap(err, "failed to save in redis repository")
	}

	input.Tz = tz8
	input.TimeLoc = int8ToLoc(tz8)

	err = r.replyIfNotSilentLocalizedf(ctx, update, input, resource.CommandTzSuccess, input.TimeLoc.String())
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
