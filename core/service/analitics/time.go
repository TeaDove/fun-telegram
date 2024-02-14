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
	"github.com/wcharczuk/go-chart/v2"
	"golang.org/x/exp/maps"
)

func getChart() chart.Chart {
	return chart.Chart{
		Width:  1000,
		Height: 1000,
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
	}
}

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

func (r *Service) getChatTimeDistribution(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsRepostChan chan<- statsRepost,
	messages []ch_repository.Message,
	tz int,
) {
	defer wg.Done()
	statsReport := statsRepost{
		repostImage: RepostImage{
			Name: "ChatTimeDistribution",
		},
	}
	today := time.Now().UTC()

	timeToCount := make(map[time.Time]float64, 100)
	for _, message := range messages {
		message.CreatedAt = message.CreatedAt.Add(time.Hour * time.Duration(tz))
		messageCreatedAt := time.Date(
			today.Year(),
			today.Month(),
			today.Day(),
			message.CreatedAt.Hour(),
			message.CreatedAt.Minute()/30*30,
			0,
			0,
			today.Location(),
		)

		timeToCount[messageCreatedAt]++
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Messages in chat by date",
			XLabel: "Date",
			YLabel: "Amount of messages",
		},
		Values:   timeToCount,
		OnlyTime: true,
	})
	if err != nil {
		statsReport.err = errors.Wrap(err, "failed to draw image in ds supplier")
		statsRepostChan <- statsReport

		return
	}

	statsReport.repostImage.Content = jpgImg
	statsRepostChan <- statsReport
}

func (r *Service) getChatDateDistribution(
	ctx context.Context,
	wg *sync.WaitGroup,
	statsReportChan chan<- statsRepost,
	messages []ch_repository.Message,
) {
	defer wg.Done()
	statsReport := statsRepost{
		repostImage: RepostImage{
			Name: "ChatDateDistribution",
		},
	}

	timeToCount := make(map[time.Time]float64, 100)
	for _, message := range messages {
		messageDate := time.Date(
			message.CreatedAt.Year(),
			message.CreatedAt.Month(),
			message.CreatedAt.Day()/7*7,
			0,
			0,
			0,
			0,
			message.CreatedAt.Location(),
		)

		timeToCount[messageDate]++
	}

	jpgImg, err := r.dsSupplier.DrawTimeseries(ctx, &ds_supplier.DrawTimeseriesInput{
		DrawInput: ds_supplier.DrawInput{
			Title:  "Messages in chat by date",
			XLabel: "Date",
			YLabel: "Amount of messages",
		},
		Values: timeToCount,
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
