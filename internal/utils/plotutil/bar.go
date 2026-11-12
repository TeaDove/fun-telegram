package plotutil

import (
	"cmp"
	"image/color"
	"math"
	"slices"
	"strconv"

	"github.com/cockroachdb/errors"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/text"
	"gonum.org/v1/plot/vg"
)

const categoryAngle = math.Pi / 6

type Bar struct {
	Label string
	Value float64
}

func (renderer Renderer) Bars(labels Labels, values []Bar, ascending bool) ([]byte, error) {
	if len(values) == 0 {
		return nil, errors.New("no bar values")
	}

	bars := slices.Clone(values)
	slices.SortFunc(bars, func(left, right Bar) int {
		order := cmp.Compare(left.Value, right.Value)
		if order == 0 {
			return cmp.Compare(left.Label, right.Label)
		}

		if !ascending {
			return -order
		}

		return order
	})

	chart := renderer.newPlot(labels, false)
	width, height, left := barCanvas(chart.X.Tick.Label, bars)
	barWidth := (width - left - vg.Points(100)) * 0.80 / vg.Length(len(bars))

	err := renderer.addBars(chart, bars, barWidth)
	if err != nil {
		return nil, err
	}

	return encodeJPEG(chart, width, height, left)
}

func (renderer Renderer) addBars(chart *plot.Plot, bars []Bar, width vg.Length) error {
	palette := [...]color.RGBA{
		{R: 114, G: 182, B: 161, A: 255}, {R: 233, G: 150, B: 117, A: 255},
		{R: 149, G: 163, B: 195, A: 255}, {R: 219, G: 149, B: 192, A: 255},
		{R: 166, G: 202, B: 106, A: 255}, {R: 241, G: 214, B: 74, A: 255},
		{R: 223, G: 196, B: 148, A: 255}, {R: 179, G: 179, B: 179, A: 255},
	}

	names := make([]string, len(bars))
	points := make(plotter.XYs, len(bars))
	numbers := make([]string, len(bars))

	for index, item := range bars {
		bar, err := plotter.NewBarChart(plotter.Values{item.Value}, width)
		if err != nil {
			return errors.Wrap(err, "create chart bar")
		}

		bar.XMin = float64(index)
		bar.Color = palette[index%len(palette)]
		bar.LineStyle.Width = 0
		chart.Add(bar)

		names[index] = item.Label
		points[index] = plotter.XY{X: float64(index), Y: item.Value}
		numbers[index] = barLabel(item.Value)
	}

	chart.NominalX(names...)
	chart.Y.Padding = 0
	chart.X.Min, chart.X.Max = -0.5, float64(len(bars))-0.5
	chart.X.Tick.Label.Rotation = categoryAngle
	chart.X.Tick.Label.XAlign = text.XRight
	chart.X.Tick.Label.YAlign = text.YTop
	chart.Y.Tick.Label.Color = color.Transparent
	chart.Y.Min, chart.Y.Max = 0, max(chart.Y.Max*1.08, 1)

	labels, err := plotter.NewLabels(plotter.XYLabels{XYs: points, Labels: numbers})
	if err != nil {
		return errors.Wrap(err, "create chart bar labels")
	}

	for index := range labels.TextStyle {
		renderer.applyFont(&labels.TextStyle[index], 11)
		labels.TextStyle[index].XAlign = text.XCenter
	}

	labels.Offset.Y = vg.Points(8)
	chart.Add(labels)

	return nil
}

func barCanvas(style text.Style, bars []Bar) (vg.Length, vg.Length, vg.Length) {
	var left, bottom vg.Length

	step := (baseWidth - vg.Points(100)) / vg.Length(len(bars))
	for index, bar := range bars {
		labelWidth := style.Width(bar.Label)
		distance := vg.Points(45) + (vg.Length(index)+0.5)*step
		left = max(left, labelWidth*vg.Length(math.Cos(categoryAngle))-distance)
		bottom = max(bottom, labelWidth*vg.Length(math.Sin(categoryAngle))-vg.Points(35))
	}

	left = min(left, vg.Points(180))
	bottom = min(bottom, vg.Points(114))

	return baseWidth + left, baseHeight + bottom, left
}

func barLabel(value float64) string {
	precision := 2
	if value == math.Trunc(value) {
		precision = 0
	}

	return strconv.FormatFloat(value, 'f', precision, 64)
}
