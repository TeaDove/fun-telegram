package analitics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/teadove/fun_telegram/core/shared"

	"github.com/teadove/fun_telegram/core/service/resource"

	"github.com/teadove/fun_telegram/core/supplier/ds_supplier"

	"github.com/pkg/errors"
)

func (r *Service) getMessagesGroupedByDateByChatId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: File{
			Name:      "MessagesGroupedByDateByChatId",
			Extension: "jpeg",
		},
	}

	messagesGrouped, err := r.dbRepository.MessageGroupByDateAndChatId(
		ctx,
		input.TgChatId,
		time.Hour*24*7,
	)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReportResult

		return
	}

	timeToCount := make(map[time.Time]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt] = float64(message.WordsCount)
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  r.resourceService.Localize(ctx, resource.AnaliseChartWordsByDate, input.Locale),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartDate, input.Locale),
			YLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartWordsWritten,
				input.Locale,
			),
		},
		Values: map[string]map[time.Time]float64{"day": timeToCount},
	})
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsReportChan <- statsReportResult

		return
	}

	statsReportResult.repostImage.Content = jpgImg
	statsReportChan <- statsReportResult
}

func (r *Service) getMessagesGroupedByDateByChatIdByUserId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: File{
			Name:      "MessagesGroupedByDateByChatIdByUserId",
			Extension: "jpeg",
		},
	}

	messagesGrouped, err := r.dbRepository.MessageGroupByDateAndChatIdAndUserId(
		ctx,
		input.TgChatId,
		input.TgUserId,
		time.Hour*24*7,
	)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReportResult

		return
	}

	timeToCount := make(map[time.Time]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt] = float64(message.WordsCount)
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  r.resourceService.Localize(ctx, resource.AnaliseChartWordsByDate, input.Locale),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartDate, input.Locale),
			YLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartWordsWritten,
				input.Locale,
			),
		},
		Values: map[string]map[time.Time]float64{"day": timeToCount},
	})
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsReportChan <- statsReportResult

		return
	}

	statsReportResult.repostImage.Content = jpgImg
	statsReportChan <- statsReportResult
}

func (r *Service) getMessagesGroupedByTimeByChatId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: File{
			Name:      "MessagesGroupedByTimeByChatId",
			Extension: "jpeg",
		},
	}

	messagesGrouped, err := r.dbRepository.MessageGroupByTimeAndChatId(
		ctx,
		input.TgChatId,
		time.Minute*10,
		input.Tz,
	)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReportResult

		return
	}
	isweekendToString := map[bool]string{
		true:  r.resourceService.Localize(ctx, resource.AnaliseChartIsWeekend, input.Locale),
		false: r.resourceService.Localize(ctx, resource.AnaliseChartIsWeekday, input.Locale),
	}

	timeToCount := make(map[string]map[time.Time]float64, 2)

	for _, message := range messagesGrouped {
		weekday := isweekendToString[message.IsWeekend]
		_, ok := timeToCount[weekday]
		if ok {
			timeToCount[weekday][message.CreatedAt] = float64(message.WordsCount)
		} else {
			timeToCount[weekday] = map[time.Time]float64{message.CreatedAt: float64(message.WordsCount)}
		}
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title: fmt.Sprintf(
				"%s UTC%s",
				r.resourceService.Localize(
					ctx,
					resource.AnaliseChartWordsByTimeOfDay,
					input.Locale,
				),
				shared.IntToSignedString(input.Tz),
			),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartTime, input.Locale),
			YLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartWordsWritten,
				input.Locale,
			),
		},
		Values:   timeToCount,
		OnlyTime: true,
	})
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsReportChan <- statsReportResult

		return
	}

	statsReportResult.repostImage.Content = jpgImg
	statsReportChan <- statsReportResult
}

func (r *Service) getMessagesGroupedByTimeByChatIdByUserId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	input *AnaliseChatInput,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: File{
			Name:      "ChatDateDistribution",
			Extension: "jpeg",
		},
	}

	messagesGrouped, err := r.dbRepository.MessageGroupByTimeAndChatIdAndUserId(
		ctx,
		input.TgChatId,
		input.TgUserId,
		time.Minute*10,
		input.Tz,
	)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReportResult

		return
	}

	isweekendToString := map[bool]string{
		true:  r.resourceService.Localize(ctx, resource.AnaliseChartIsWeekend, input.Locale),
		false: r.resourceService.Localize(ctx, resource.AnaliseChartIsWeekday, input.Locale),
	}

	timeToCount := make(map[string]map[time.Time]float64, 2)

	for _, message := range messagesGrouped {
		weekday := isweekendToString[message.IsWeekend]
		_, ok := timeToCount[weekday]
		if ok {
			timeToCount[weekday][message.CreatedAt] = float64(message.WordsCount)
		} else {
			timeToCount[weekday] = map[time.Time]float64{message.CreatedAt: float64(message.WordsCount)}
		}
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title: fmt.Sprintf(
				"%s UTC%s",
				r.resourceService.Localize(
					ctx,
					resource.AnaliseChartWordsByTimeOfDay,
					input.Locale,
				),
				shared.IntToSignedString(input.Tz),
			),
			XLabel: r.resourceService.Localize(ctx, resource.AnaliseChartTime, input.Locale),
			YLabel: r.resourceService.Localize(
				ctx,
				resource.AnaliseChartWordsWritten,
				input.Locale,
			),
		},
		Values:   timeToCount,
		OnlyTime: true,
	})
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsReportChan <- statsReportResult

		return
	}

	statsReportResult.repostImage.Content = jpgImg
	statsReportChan <- statsReportResult
}
