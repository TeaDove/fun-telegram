package resource

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog"
	"github.com/teadove/goteleout/internal/utils"
)

type Service struct {
	Locales mapset.Set[Locale]
}

func New(ctx context.Context) (*Service, error) {
	r := Service{}
	r.Locales = locales

	return &r, nil
}

func (r *Service) Localize(ctx context.Context, code Code, locale Locale) string {
	text, ok := localizer[code][locale]
	if !ok {
		zerolog.Ctx(ctx).Warn().
			Str("status", "localized.string.not.found").
			Int("code", int(code)).
			Str("locale", string(locale)).
			Send()
		return utils.Undefined
	}

	return text
}

func (r *Service) Localizef(ctx context.Context, code Code, locale Locale, args ...any) string {
	return fmt.Sprintf(r.Localize(ctx, code, locale), args...)
}
