package analitics

import (
	"context"
	"strings"
	"sync"

	"github.com/teadove/fun_telegram/core/service/resource"

	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
)

func (r *Service) getMostToxicUsers(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
	getter nameGetter,
) {
	defer wg.Done()
	const maxUsers = 15
	output := statsReport{
		repostImage: File{
			Name:      "MostToxicUsers",
			Extension: "jpeg",
		},
	}

	userToCountArray, err := r.chRepository.GroupedCountGetByChatIdByUserId(ctx, input.TgChatId, maxUsers)
	if err != nil {
		output.err = errors.Wrap(err, "failed to get GroupedCountGetByChatIdByUserId")
		statsReportChan <- output

		return
	}

	userToCount := make(map[string]float64, maxUsers)
	for _, message := range userToCountArray {
		userToCount[getter.GetName(message.TgUserId)] = float64(message.ToxicWordsCount) / float64(message.WordsCount) * 100
	}

	jpgImg, err := r.dsSupplier.DrawBar(ctx, &ds_supplier.DrawBarInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  r.resourceService.Localize(ctx, resource.AnaliseChartToxicityPercentShort, input.Locale),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartUser, input.Locale),
			YLabel: r.resourceService.Localize(ctx, resource.AnaliseChartToxicityPercentLong, input.Locale),
		},
		Values: userToCount,
	})
	if err != nil {
		output.err = errors.Wrap(err, "failed to draw in ds supplier")
		statsReportChan <- output

		return
	}

	output.repostImage.Content = jpgImg
	statsReportChan <- output
}

func (r *Service) IsToxic(word string) (bool, error) {
	match, err := r.toxicityExp.MatchString(word)
	if err != nil {
		return false, errors.WithStack(err)
	}

	return match, nil
}

func (r *Service) IsToxicSentence(words string) (string, bool, error) {
	sentence := strings.Fields(strings.TrimSpace(strings.ToLower(words)))
	for _, word := range sentence {
		match, err := r.toxicityExp.MatchString(word)
		if err != nil {
			return "", false, errors.WithStack(err)
		}

		if match {
			return word, true, nil
		}
	}

	return "", false, nil
}
