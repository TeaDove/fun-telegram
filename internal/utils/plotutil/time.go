package plotutil

import (
	"cmp"
	"image/color"
	"math"
	"slices"
	"time"

	"github.com/cockroachdb/errors"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const day = 24 * time.Hour

type TimePoint struct {
	Time  time.Time
	Value float64
}

func (renderer Renderer) TimeSeries(labels Labels, legend string, values []TimePoint) ([]byte, error) {
	if len(values) == 0 {
		return nil, errors.New("no time series values")
	}

	points := make(plotter.XYs, len(values))
	for index, value := range values {
		points[index] = plotter.XY{X: float64(value.Time.Unix()), Y: value.Value}
	}

	slices.SortFunc(points, func(left, right plotter.XY) int {
		return cmp.Compare(left.X, right.X)
	})

	line, markers, err := plotter.NewLinePoints(points)
	if err != nil {
		return nil, errors.Wrap(err, "create time series line")
	}

	seriesColor := color.RGBA{R: 102, G: 194, B: 165, A: 255}
	line.Color = seriesColor
	line.Width = vg.Points(1.6)
	markers.Color = seriesColor
	markers.Shape = draw.CircleGlyph{}
	markers.Radius = vg.Points(3)

	chart := renderer.newPlot(labels, true)
	chart.Add(line, markers)
	chart.Legend.Add(legend, line, markers)
	chart.Legend.Top = true
	chart.Legend.XOffs = -vg.Points(8)
	chart.Legend.YOffs = -vg.Points(8)

	daySeconds := day.Seconds()
	timeMargin := max((chart.X.Max-chart.X.Min)*0.05, daySeconds)
	chart.X.Min -= timeMargin
	chart.X.Max += timeMargin
	valueMargin := max((chart.Y.Max-chart.Y.Min)*0.05, 1)
	chart.Y.Min -= valueMargin
	chart.Y.Max += valueMargin
	location := values[0].Time.Location()
	chart.X.Tick.Marker = plot.TickerFunc(func(minimum, maximum float64) []plot.Tick {
		return dateTicks(minimum, maximum, location)
	})

	return encodeJPEG(chart, baseWidth, 10*vg.Inch, 0)
}

func dateTicks(minimum, maximum float64, location *time.Location) []plot.Tick {
	start := time.Unix(int64(minimum), 0).In(location)
	end := time.Unix(int64(maximum), 0).In(location)
	days := end.Sub(start).Hours() / 24
	date := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, location)
	dayStep := max(1, int(math.Ceil(days/8)))

	var monthStep int

	format := time.DateOnly

	if days >= 120 {
		date = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, location)
		monthStep = max(1, int(math.Ceil(days/(30*8))))
		dayStep = 0
		format = "2006-01"
	}

	ticks := make([]plot.Tick, 0, 10)

	for !date.After(end) {
		if !date.Before(start) {
			ticks = append(ticks, plot.Tick{Value: float64(date.Unix()), Label: date.Format(format)})
		}

		date = date.AddDate(0, monthStep, dayStep)
	}

	return ticks
}
