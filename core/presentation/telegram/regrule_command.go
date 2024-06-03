package telegram

import (
	"fmt"
	"strconv"
	"time"

	"github.com/celestix/gotgproto/ext"
	"github.com/dlclark/regexp2"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/fun_telegram/core/service/resource"
)

var (
	FlagRegRuleDelete = optFlag{
		Long:        "delete",
		Short:       "d",
		Description: resource.CommandKandinskyFlagNegativePromptDescription,
	}
	FlagRegRuleRegexp = optFlag{
		Long:        "regexp",
		Short:       "r",
		Description: resource.CommandKandinskyFlagStyleDescription,
	}
	FlagRegRuleList = optFlag{
		Long:        "list",
		Short:       "l",
		Description: resource.CommandKandinskyFlagStyleDescription,
	}
)

func (r *Presentation) regruleCommandList(
	ctx *ext.Context,
	update *ext.Update,
) error {
	rules, err := r.redisRepository.GetRegRules(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.Wrap(err, "failed to get rules")
	}

	text := make([]styling.StyledTextOption, 0, 4)
	text = append(text, styling.Plain("Regexp rules: \n\n"))
	idx := 0
	for reg, rule := range rules {
		text = append(text,
			styling.Plain(strconv.Itoa(idx)),
			styling.Plain(") "),
			styling.Code(reg),
			styling.Plain(": "),
			styling.Plain(rule),
			styling.Plain("\n\n"),
		)
		idx++
	}

	_, err = ctx.Reply(update, text, nil)
	if err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	return nil
}

func (r *Presentation) regruleCommandHandler(
	ctx *ext.Context,
	update *ext.Update,
	input *input,
) error {
	if _, ok := input.Ops[FlagRegRuleList.Long]; ok {
		return r.regruleCommandList(ctx, update)
	}

	if regToDelete, ok := input.Ops[FlagRegRuleDelete.Long]; ok {
		err := r.redisRepository.DelRegRules(ctx, update.EffectiveChat().GetID(), regToDelete)
		if err != nil {
			return errors.Wrap(err, "failed to delete")
		}

		_, err = ctx.Reply(update, "Ok", nil)
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}

		return nil
	}

	regToSet, ok := input.Ops[FlagRegRuleRegexp.Long]
	if !ok {
		err := r.replyIfNotSilentLocalizedf(
			ctx,
			update,
			input,
			resource.ErrUnprocessableEntity,
			"--regexp required, but not passed",
		)
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}

		return nil
	}

	reg, err := regexp2.Compile(regToSet, 0)
	if err != nil {
		err = r.replyIfNotSilentLocalizedf(
			ctx,
			update,
			input,
			resource.ErrUnprocessableEntity,
			fmt.Sprintf("bad regexp: %s", err.Error()),
		)
		if err != nil {
			return errors.Wrap(err, "failed to reply")
		}

		return nil
	}

	err = r.redisRepository.SetRegRules(
		ctx,
		update.EffectiveChat().GetID(),
		reg.String(),
		input.Text,
	)
	if err != nil {
		return errors.Wrap(err, "failed to set reg rule")
	}

	_, err = ctx.Reply(
		update,
		[]styling.StyledTextOption{styling.Plain("Ok: "), styling.Code(reg.String())},
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "failed to reply")
	}

	return nil
}

func (r *Presentation) regRuleFinderMessagesProcessor(ctx *ext.Context, update *ext.Update) error {
	chatSettings, err := r.getChatSettings(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.WithStack(err)
	}

	if !chatSettings.Enabled {
		return nil
	}

	ok := r.checkFeatureEnabled(&chatSettings, "regrule")
	if !ok {
		return nil
	}

	regRules, err := r.redisRepository.GetRegRules(ctx, update.EffectiveChat().GetID())
	if err != nil {
		return errors.Wrap(err, "failed to get reg rules")
	}

	if len(regRules) == 0 {
		return nil
	}

	for reg, rule := range regRules {
		regexp, err := regexp2.Compile(reg, regexp2.IgnoreCase)
		if err != nil {
			zerolog.Ctx(ctx).
				Error().
				Stack().
				Err(err).
				Str("status", "bad.regexp").
				Str("regexp", reg).
				Send()

			continue
		}

		regexp.MatchTimeout = time.Second * 2

		match, err := regexp.MatchString(update.EffectiveMessage.Text)
		if err != nil {
			zerolog.Ctx(ctx).
				Error().
				Stack().
				Err(err).
				Str("status", "failed.to.match").
				Str("regexp", reg).
				Send()
			continue
		}

		if !match {
			continue
		}

		if match {
			_, err = ctx.Reply(update, rule, nil)
			if err != nil {
				return errors.Wrap(err, "failed to reply")
			}
		}
	}

	return nil
}
