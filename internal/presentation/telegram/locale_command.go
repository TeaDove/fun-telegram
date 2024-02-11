package telegram

import (
	"fmt"
	"github.com/celestix/gotgproto/ext"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/redis_repository"
	"github.com/teadove/goteleout/internal/service/resource"
	"strings"
)

func getLocalePath(chatId int64) string {
	return fmt.Sprintf("locale::%d", chatId)
}

const defaultLocale = resource.En

func (r *Presentation) getLocale(chatId int64) (resource.Locale, error) {
	localeBytes, err := r.redisRepository.Load(getLocalePath(chatId))
	if err != nil {
		if errors.Is(err, redis_repository.ErrKeyNotFound) {
			return defaultLocale, nil
		}
		return "", errors.WithStack(err)
	}

	return resource.Locale(localeBytes), nil
}

func (r *Presentation) localeCommandHandler(ctx *ext.Context, update *ext.Update, input *Input) error {
	locale := resource.Locale(strings.ToLower(strings.TrimSpace(input.Text)))
	if !r.resourceService.Locales.Contains(locale) {
		err := r.replyIfNotSilent(ctx, update, input, r.resourceService.Localizef(ctx, resource.ErrLocaleNotFound, input.Locale, locale))
		if err != nil {
			return errors.WithStack(err)
		}

		return nil
	}

	path := getLocalePath(update.EffectiveChat().GetID())
	err := r.redisRepository.Save(path, []byte(locale))
	if err != nil {
		return errors.WithStack(err)
	}

	input.Locale = locale

	err = r.replyIfNotSilentLocalized(ctx, update, input, resource.CommandLocaleSuccess)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}