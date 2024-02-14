package analitics

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/teadove/goteleout/core/supplier/ds_supplier"

	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/repository/ch_repository"
	"golang.org/x/exp/maps"
)

func getChatterBoxes(messages []ch_repository.Message, maxUsers int) ([]int64, map[int64]int) {
	userToCount := make(map[int64]int, 100)
	for _, message := range messages {
		wordsCount := len(strings.Fields(message.Text))
		_, ok := userToCount[message.TgUserId]
		if ok {
			userToCount[message.TgUserId] += wordsCount
		} else {
			userToCount[message.TgUserId] = wordsCount
		}
	}

	users := maps.Keys(userToCount)
	sort.SliceStable(users, func(i, j int) bool {
		return userToCount[users[i]] > userToCount[users[j]]
	})
	if len(users) > maxUsers {
		users = users[:maxUsers]
	}

	return users, userToCount
}

func (r *Service) getMessagesGroupedByDateByChatId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsRepostChan chan<- statsReport,
	chatId int64,
) {
	defer wg.Done()
	statsReport := statsReport{
		repostImage: RepostImage{
			Name: "MessagesGroupedByDateByChatId",
		},
	}
	messagesGrouped, err := r.chRepository.GetMessagesGroupedByDateByChatId(ctx, chatId, 86400*7)
	if err != nil {
		statsReport.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsRepostChan <- statsReport

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
		Values: timeToCount,
	})
	if err != nil {
		statsReport.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsRepostChan <- statsReport

		return
	}

	statsReport.repostImage.Content = jpgImg
	statsRepostChan <- statsReport
}

func (r *Service) getMessagesGroupedByDateByChatIdByUserId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsRepostChan chan<- statsReport,
	chatId int64,
	userId int64,
) {
	defer wg.Done()
	statsReport := statsReport{
		repostImage: RepostImage{
			Name: "MessagesGroupedByDateByChatIdByUserId",
		},
	}
	messagesGrouped, err := r.chRepository.GetMessagesGroupedByDateByChatIdByUserId(ctx, chatId, userId, 86400*7)
	if err != nil {
		statsReport.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsRepostChan <- statsReport

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
		Values: timeToCount,
	})
	if err != nil {
		statsReport.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsRepostChan <- statsReport

		return
	}

	statsReport.repostImage.Content = jpgImg
	statsRepostChan <- statsReport
}

func (r *Service) getMessagesGroupedByTimeByChatId(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsReport,
	chatId int64,
	tz int,

) {
	defer wg.Done()
	statsReport := statsReport{
		repostImage: RepostImage{
			Name: "MessagesGroupedByTimeByChatId",
		},
	}

	messagesGrouped, err := r.chRepository.GetMessagesGroupedByTimeByChatId(ctx, chatId, 60*30)
	if err != nil {
		statsReport.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReport

		return
	}

	timeToCount := make(map[time.Time]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt] = float64(message.Count)
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
		statsReport.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsReportChan <- statsReport

		return
	}

	statsReport.repostImage.Content = jpgImg
	statsReportChan <- statsReport
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
	statsReport := statsReport{
		repostImage: RepostImage{
			Name: "ChatDateDistribution",
		},
	}

	messagesGrouped, err := r.chRepository.GetMessagesGroupedByTimeByChatIdByUserId(ctx, chatId, userId, 60*30)
	if err != nil {
		statsReport.err = errors.Wrap(err, "failed to get messages from ch repository")
		statsReportChan <- statsReport

		return
	}

	timeToCount := make(map[time.Time]float64, 100)
	for _, message := range messagesGrouped {
		timeToCount[message.CreatedAt] = float64(message.Count)
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
		statsReport.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsReportChan <- statsReport

		return
	}

	statsReport.repostImage.Content = jpgImg
	statsReportChan <- statsReport

	return
}
