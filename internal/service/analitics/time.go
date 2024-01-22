package analitics

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/internal/repository/db_repository"
	"github.com/wcharczuk/go-chart/v2"
	"golang.org/x/exp/maps"
	"sort"
	"time"
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

func (r *Service) getChatTimeDistributionByUser(messages []db_repository.Message, getter nameGetter, tz int) ([]byte, error) {
	const (
		minuteRate = 30
		maxUsers   = 10
	)
	timeToCount := make(map[int64]map[float64]int, 100)
	userToMessages := make(map[int64]int, 100)
	for _, message := range messages {
		message.CreatedAt = message.CreatedAt.Add(time.Hour * time.Duration(tz))
		messageTime := float64(message.CreatedAt.Hour()) + float64(message.CreatedAt.Minute()/minuteRate*minuteRate)/60
		userId := message.TgUserId

		timeMap, ok := timeToCount[userId]
		if ok {
			_, ok = timeMap[messageTime]
			if ok {
				timeMap[messageTime]++
			} else {
				timeMap[messageTime] = 1
			}
		} else {
			timeToCount[userId] = map[float64]int{messageTime: 1}
		}

		_, ok = userToMessages[userId]
		if ok {
			userToMessages[userId]++
		} else {
			userToMessages[userId] = 1
		}
	}

	chartDrawn := getChart()
	chartDrawn.Title = fmt.Sprintf("Message count distribution by time UTC+%d", tz)
	chartDrawn.Series = make([]chart.Series, 0, maxUsers)

	users := maps.Keys(userToMessages)
	sort.SliceStable(users, func(i, j int) bool {
		return users[i] > users[j]
	})
	if len(users) > maxUsers {
		users = users[:maxUsers]
	}

	for _, userId := range users {
		var values chart.ContinuousSeries
		values.Name = getter.Get(userId)
		values.XValues = make([]float64, 0, len(timeToCount))
		values.YValues = make([]float64, 0, len(timeToCount))

		times := maps.Keys(timeToCount[userId])
		sort.SliceStable(times, func(i, j int) bool {
			return times[i] > times[j]
		})

		for _, chatTime := range times {
			values.XValues = append(values.XValues, chatTime)
			values.YValues = append(values.YValues, float64(timeToCount[userId][chatTime]))
		}

		chartDrawn.Series = append(chartDrawn.Series, values)
	}

	chartDrawn.Elements = []chart.Renderable{chart.Legend(&chartDrawn)}

	var chartBuffer bytes.Buffer

	err := chartDrawn.Render(chart.PNG, &chartBuffer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jpgImg, err := PngToJpeg(chartBuffer.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return jpgImg, nil
}

func (r *Service) getChatTimeDistribution(messages []db_repository.Message, tz int) ([]byte, error) {
	const minuteRate = 30
	timeToCount := make(map[float64]int, 100)
	for _, message := range messages {
		message.CreatedAt = message.CreatedAt.Add(time.Hour * time.Duration(tz))
		messageTime := float64(message.CreatedAt.Hour()) + float64(message.CreatedAt.Minute()/minuteRate*minuteRate)/60
		_, ok := timeToCount[messageTime]
		if ok {
			timeToCount[messageTime]++
		} else {
			timeToCount[messageTime] = 1
		}
	}

	times := maps.Keys(timeToCount)
	sort.SliceStable(times, func(i, j int) bool {
		return times[i] > times[j]
	})

	var values chart.ContinuousSeries
	values.XValues = make([]float64, 0, len(timeToCount))
	values.YValues = make([]float64, 0, len(timeToCount))

	for _, chatTime := range times {
		values.XValues = append(values.XValues, chatTime)
		values.YValues = append(values.YValues, float64(timeToCount[chatTime]))
	}

	chartDrawn := getChart()
	chartDrawn.Title = fmt.Sprintf("Message count distribution by time UTC+%d", tz)
	chartDrawn.Series = []chart.Series{values}

	var chartBuffer bytes.Buffer

	err := chartDrawn.Render(chart.PNG, &chartBuffer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jpgImg, err := PngToJpeg(chartBuffer.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return jpgImg, nil
}

func (r *Service) getChatDateDistribution(messages []db_repository.Message) ([]byte, error) {
	timeToCount := make(map[time.Time]int, 100)
	for _, message := range messages {
		messageDate := time.Date(message.CreatedAt.Year(), message.CreatedAt.Month(), message.CreatedAt.Day()/3*3, 0, 0, 0, 0, message.CreatedAt.Location())

		_, ok := timeToCount[messageDate]
		if ok {
			timeToCount[messageDate]++
		} else {
			timeToCount[messageDate] = 1
		}
	}

	times := maps.Keys(timeToCount)
	sort.SliceStable(times, func(i, j int) bool {
		return times[i].After(times[j])
	})

	var values chart.TimeSeries
	values.XValues = make([]time.Time, 0, len(timeToCount))
	values.YValues = make([]float64, 0, len(timeToCount))

	for _, chatTime := range times {
		values.XValues = append(values.XValues, chatTime)
		values.YValues = append(values.YValues, float64(timeToCount[chatTime]))
	}

	chartDrawn := getChart()
	chartDrawn.Title = "Message count distribution by date"
	chartDrawn.Series = []chart.Series{values}

	var chartBuffer bytes.Buffer

	err := chartDrawn.Render(chart.PNG, &chartBuffer)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	jpgImg, err := PngToJpeg(chartBuffer.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return jpgImg, nil
}
