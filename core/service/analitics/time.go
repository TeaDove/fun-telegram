package analitics

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

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

func (r *Service) getChatTimeDistribution(messages []ch_repository.Message, tz int) ([]byte, error) {
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

func (r *Service) getChatDateDistribution(messages []ch_repository.Message) ([]byte, error) {
	timeToCount := make(map[time.Time]int, 100)
	for _, message := range messages {
		messageDate := time.Date(
			message.CreatedAt.Year(),
			message.CreatedAt.Month(),
			message.CreatedAt.Day()/3*3,
			0,
			0,
			0,
			0,
			message.CreatedAt.Location(),
		)

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

	if len(values.XValues) < 4 {
		return nil, nil
	}

	chartDrawn := getChart()
	chartDrawn.Title = "Message count distribution by date"
	chartDrawn.Series = []chart.Series{values, &chart.PolynomialRegressionSeries{
		Degree:      2,
		InnerSeries: values,
		Name:        "PolynomialRegression",
	}}
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
