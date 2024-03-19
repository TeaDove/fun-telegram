package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/celestix/gotgproto/ext"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/repository/redis_repository"

	"github.com/teadove/fun_telegram/core/service/resource"
)

type ChatSettings struct {
	Enabled  bool            `json:"enabled"`
	Locale   resource.Locale `json:"locale"`
	Tz       int8            `json:"tz"`
	TimeLoc  *time.Location  `json:"timeLoc"`
	Features map[string]bool `json:"features"`
}

const defaultLocale = resource.En

func (r *Presentation) getDefaultSettings() ChatSettings {
	settings := ChatSettings{
		Enabled: false,
		Locale:  defaultLocale,
		Tz:      0,
		TimeLoc: time.UTC,
	}

	return settings
}

func (r *Presentation) getChatSettings(ctx context.Context, chatId int64) (ChatSettings, error) {
	redisChatSettings, err := r.redisRepository.GetChatSettings(ctx, chatId)
	if err != nil {
		if errors.Is(err, redis_repository.ErrKeyNotFound) {
			return r.getDefaultSettings(), nil
		}

		return ChatSettings{}, errors.Wrap(err, "failed to get chat settings")
	}

	var setChannelLocale bool

	if redisChatSettings.Locale == "" {
		zerolog.Ctx(ctx).
			Warn().
			Str("status", "empty.locale").
			Interface("chat", redisChatSettings).
			Send()
		redisChatSettings.Locale = string(defaultLocale)
		setChannelLocale = true
	}

	chatSettings := ChatSettings{
		Enabled:  redisChatSettings.Enabled,
		Locale:   resource.Locale(redisChatSettings.Locale),
		Tz:       redisChatSettings.Tz,
		TimeLoc:  int8ToLoc(redisChatSettings.Tz),
		Features: make(map[string]bool, 3),
	}

	if len(redisChatSettings.Features) != 0 {
		err = json.Unmarshal(redisChatSettings.Features, &chatSettings.Features)
		if err != nil {
			return ChatSettings{}, errors.Wrap(err, "failed to unmarshal features")
		}
	}

	if setChannelLocale {
		err = r.setChatSettings(ctx, chatId, &chatSettings)
		if err != nil {
			return ChatSettings{}, errors.Wrap(
				err,
				"failed to set channel settings because of failed parsing",
			)
		}
	}

	return chatSettings, nil
}

func (r *Presentation) setChatSettings(
	ctx context.Context,
	chatId int64,
	newChatSettings *ChatSettings,
) (err error) {
	redisChatSettings := redis_repository.ChatSettings{
		Enabled:  newChatSettings.Enabled,
		Locale:   string(newChatSettings.Locale),
		Tz:       newChatSettings.Tz,
		Features: nil,
	}
	if len(newChatSettings.Features) != 0 {
		redisChatSettings.Features, err = json.Marshal(newChatSettings.Features)
		if err != nil {
			return errors.Wrap(err, "failed to marshal features")
		}
	}

	err = r.redisRepository.SetChatSettings(ctx, chatId, &redisChatSettings)
	if err != nil {
		return errors.Wrap(err, "failed to set chat settings")
	}

	return nil
}

func (r *Presentation) checkFeatureEnabled(chatSettings *ChatSettings, featureName string) bool {
	featureToggled, ok := chatSettings.Features[featureName]
	if ok {
		return featureToggled
	}

	return r.features[featureName]
}

var (
	FlagChatEnabled = optFlag{
		Long:        "enabled",
		Short:       "e",
		Description: resource.CommandChatFlagEnableDescription,
	}
	FlagChatLocale = optFlag{
		Long:        "locale",
		Short:       "l",
		Description: resource.CommandChatFlagLocaleDescription,
	}
	FlagChatTz = optFlag{
		Long:        "timezone",
		Short:       "t",
		Description: resource.CommandChatFlagTzDescription,
	}
)

// TODO fix nocyclo

// chatCommandHandler
// nolint: gocyclo
func (r *Presentation) chatCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) (err error) {
	chatId := update.EffectiveChat().GetID()
	msgText := make([]styling.StyledTextOption, 0, 5)

	enable, ok := input.Ops[FlagChatEnabled.Long]
	if ok {
		if enable == "" {
			if input.ChatSettings.Enabled {
				msgText = append(msgText, styling.Plain("Bot disabled in this chat\n\n"))
				input.ChatSettings.Enabled = false
			} else {
				msgText = append(msgText, styling.Plain("Bot enabled in this chat\n\n"))
				input.ChatSettings.Enabled = true
			}
		} else {
			_, ok = r.features[enable]
			if !ok {
				err = r.replyIfNotSilentLocalized(ctx, update, input, resource.ErrCommandNotFound)
				if err != nil {
					return errors.Wrap(err, "failed to reply with silence")
				}

				return nil
			}

			if enable == "chat" {
				err = r.replyIfNotSilentLocalized(ctx, update, input, resource.ErrNiceTry)
				if err != nil {
					return errors.Wrap(err, "failed to reply with silence")
				}

				return nil
			}

			ok = r.checkFeatureEnabled(&input.ChatSettings, enable)
			if ok {
				msgText = append(msgText, styling.Plain(fmt.Sprintf("Feature disabled: %s", enable)))
				input.ChatSettings.Features[enable] = false
			} else {
				msgText = append(msgText, styling.Plain(fmt.Sprintf("Feature enabled: %s", enable)))
				input.ChatSettings.Features[enable] = true
			}
		}
	}

	locale, ok := input.Ops[FlagChatLocale.Long]

	if ok {
		resourceLocale := resource.Locale(strings.ToLower(strings.TrimSpace(locale)))
		if !r.resourceService.Locales.Contains(resourceLocale) {
			err = r.replyIfNotSilent(
				ctx,
				update,
				input,
				r.resourceService.Localizef(
					ctx,
					resource.ErrLocaleNotFound,
					input.ChatSettings.Locale,
					locale,
				),
			)
			if err != nil {
				return errors.WithStack(err)
			}

			return nil
		}

		input.ChatSettings.Locale = resourceLocale

		msgText = append(
			msgText,
			styling.Plain(
				r.resourceService.Localize(ctx, resource.CommandChatLocaleSuccess, resourceLocale),
			),
			styling.Plain("\n\n"),
		)
	}

	tz, ok := input.Ops[FlagChatTz.Long]
	if ok {
		tzInt, err := strconv.Atoi(strings.TrimSpace(tz))
		if err != nil {
			err = r.replyIfNotSilent(
				ctx,
				update,
				input,
				r.resourceService.Localizef(
					ctx,
					resource.ErrUnprocessableEntity,
					input.ChatSettings.Locale,
					err,
				),
			)
			if err != nil {
				return errors.Wrap(err, "failed to reply if not silent")
			}

			return nil
		}

		if !(-12 < tzInt && tzInt < 12) {
			err = r.replyIfNotSilent(
				ctx,
				update,
				input,
				r.resourceService.Localizef(
					ctx,
					resource.ErrUnprocessableEntity,
					input.ChatSettings.Locale,
					errors.New("tz should be from -12 to 12"),
				),
			)
			if err != nil {
				return errors.Wrap(err, "failed to reply if not silent")
			}

			return nil
		}

		tz8 := int8(tzInt)
		input.ChatSettings.Tz = tz8
		input.ChatSettings.TimeLoc = int8ToLoc(tz8)

		msgText = append(
			msgText,
			styling.Plain(
				r.resourceService.Localizef(
					ctx,
					resource.CommandChatTzSuccess,
					input.ChatSettings.Locale,
					input.ChatSettings.TimeLoc.String(),
				),
			),
		)
	}

	err = r.setChatSettings(ctx, chatId, &input.ChatSettings)
	if err != nil {
		return errors.Wrap(err, "failed to set chat settings")
	}

	if len(msgText) == 0 {
		err = r.replyIfNotSilent(ctx, update, input, "Nothing to change")
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}

		return nil
	}

	err = r.replyIfNotSilent(ctx, update, input, msgText)
	if err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	return nil
}
