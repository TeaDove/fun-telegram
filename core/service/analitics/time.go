package analitics

import (
	"context"
	"sync"
	"time"

	"github.com/teadove/goteleout/core/supplier/ds_supplier"

	"github.com/pkg/errors"
)

func (r *Service) getMessagesGroupedByDateByChatId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	chatId int64,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: RepostImage{
			Name: "MessagesGroupedByDateByChatId",
		},
	}
	messagesGrouped, err := r.chRepository.GetMessagesGroupedByDateByChatId(ctx, chatId, 86400*7)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReportResult

		return
	}

	timeToCount := make(map[time.Time]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt] = float64(message.Count)
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Messages in chat by day",
			XLabel: "Day",
			YLabel: "Amount of messages",
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
	chatId int64,
	userId int64,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: RepostImage{
			Name: "MessagesGroupedByDateByChatIdByUserId",
		},
	}
	messagesGrouped, err := r.chRepository.GetMessagesGroupedByDateByChatIdByUserId(ctx, chatId, userId, 86400*7)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReportResult

		return
	}

	timeToCount := make(map[time.Time]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt] = float64(message.Count)
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Messages in chat by day",
			XLabel: "Day",
			YLabel: "Amount of messages",
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

var isweekendToString = map[bool]string{true: "is weekend", false: "is weekday"}

func (r *Service) getMessagesGroupedByTimeByChatId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	chatId int64,
	tz int,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: RepostImage{
			Name: "MessagesGroupedByTimeByChatId",
		},
	}

	messagesGrouped, err := r.chRepository.GetMessagesGroupedByTimeByChatId(ctx, chatId, 60*30, tz)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReportResult

		return
	}

	timeToCount := make(map[string]map[time.Time]float64, 2)
	for _, message := range messagesGrouped {
		weekday := isweekendToString[message.IsWeekend]
		_, ok := timeToCount[weekday]
		if ok {
			timeToCount[weekday][message.CreatedAt] = float64(message.Count)
		} else {
			timeToCount[weekday] = map[time.Time]float64{message.CreatedAt: float64(message.Count)}
		}
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Messages in chat by time of day",
			XLabel: "Time",
			YLabel: "Amount of messages",
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
	chatId int64,
	userId int64,
	tz int,
) {
	defer wg.Done()
	statsReportResult := statsReport{
		repostImage: RepostImage{
			Name: "ChatDateDistribution",
		},
	}

	messagesGrouped, err := r.chRepository.GetMessagesGroupedByTimeByChatIdByUserId(ctx, chatId, userId, 60*30, tz)
	if err != nil {
		statsReportResult.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReportResult

		return
	}

	timeToCount := make(map[string]map[time.Time]float64, 2)
	for _, message := range messagesGrouped {
		weekday := isweekendToString[message.IsWeekend]
		_, ok := timeToCount[weekday]
		if ok {
			timeToCount[weekday][message.CreatedAt] = float64(message.Count)
		} else {
			timeToCount[weekday] = map[time.Time]float64{message.CreatedAt: float64(message.Count)}
		}
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Messages in chat by time of day",
			XLabel: "Time",
			YLabel: "Amount of messages",
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

	return
}
