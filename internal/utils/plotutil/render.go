package plotutil

import (
	"bytes"
	"image/color"
	"image/jpeg"
	"math"
	"strconv"
	"strings"

	"codeberg.org/go-fonts/dejavu/dejavusans"
	"github.com/cockroachdb/errors"
	"golang.org/x/image/font/opentype"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/text"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

const (
	imageDPI      = 300
	baseWidth     = 16.5 * vg.Inch
	baseHeight    = 10.2 * vg.Inch
	canvasPadding = 16 * vg.Length(1)
)

type Labels struct {
	Title string
	X     string
	Y     string
}

type Renderer struct {
	font    font.Font
	handler text.Plain
}

func New() (Renderer, error) {
	face, err := opentype.Parse(dejavusans.TTF)
	if err != nil {
		return Renderer{}, errors.Wrap(err, "parse chart font")
	}

	descriptor := font.Font{Typeface: "DejaVu", Variant: "Sans"}
	cache := font.NewCache(font.Collection{{Font: descriptor, Face: face}})

	return Renderer{font: descriptor, handler: text.Plain{Fonts: cache}}, nil
}

func (renderer Renderer) applyFont(style *text.Style, size float64) {
	style.Font = font.From(renderer.font, vg.Points(size))
	style.Handler = renderer.handler
	style.Color = color.Gray{Y: 38}
}

func (renderer Renderer) newPlot(labels Labels, verticalGrid bool) *plot.Plot {
	chart := plot.New()
	chart.TextHandler = renderer.handler
	chart.Title.Text = labels.Title
	chart.Title.Padding = vg.Points(6)
	chart.X.Label.Text = labels.X
	chart.Y.Label.Text = labels.Y
	renderer.applyFont(&chart.Title.TextStyle, 14)
	renderer.applyFont(&chart.X.Label.TextStyle, 14)
	renderer.applyFont(&chart.Y.Label.TextStyle, 14)
	renderer.applyFont(&chart.X.Tick.Label, 11)
	renderer.applyFont(&chart.Y.Tick.Label, 11)
	renderer.applyFont(&chart.Legend.TextStyle, 14)

	chart.X.Width, chart.Y.Width = 0, 0
	chart.X.Tick.Length, chart.Y.Tick.Length = 0, 0
	chart.Y.Tick.Marker = plot.TickerFunc(valueTicks)
	grid := plotter.NewGrid()
	grid.Horizontal = draw.LineStyle{Color: color.Gray{Y: 205}, Width: vg.Points(0.7)}

	if verticalGrid {
		grid.Vertical = grid.Horizontal
	} else {
		grid.Vertical.Color = nil
	}

	chart.Add(grid, frame{})

	return chart
}

func encodeJPEG(chart *plot.Plot, width, height, left vg.Length) ([]byte, error) {
	canvas := vgimg.NewWith(vgimg.UseWH(width, height), vgimg.UseDPI(imageDPI))
	chart.Draw(draw.Crop(draw.New(canvas), left+canvasPadding, -canvasPadding, canvasPadding, -canvasPadding))

	var buffer bytes.Buffer

	err := jpeg.Encode(&buffer, canvas.Image(), &jpeg.Options{Quality: 85})
	if err != nil {
		return nil, errors.Wrap(err, "encode chart JPEG")
	}

	jfif := []byte{
		0xff, 0xe0, 0, 16, 'J', 'F', 'I', 'F', 0, 1, 1, 1,
		byte(imageDPI >> 8), byte(imageDPI & 255), byte(imageDPI >> 8), byte(imageDPI & 255), 0, 0,
	}
	encoded := buffer.Bytes()
	result := make([]byte, 0, len(encoded)+len(jfif))
	result = append(result, encoded[:2]...)
	result = append(result, jfif...)
	result = append(result, encoded[2:]...)

	return result, nil
}

type frame struct{}

func (frame) Plot(canvas draw.Canvas, _ *plot.Plot) {
	bounds := canvas.Rectangle
	style := draw.LineStyle{Color: color.Gray{Y: 195}, Width: vg.Points(0.7)}
	canvas.StrokeLines(style, []vg.Point{
		bounds.Min,
		{X: bounds.Max.X, Y: bounds.Min.Y},
		bounds.Max,
		{X: bounds.Min.X, Y: bounds.Max.Y},
		bounds.Min,
	})
}

func valueTicks(minimum, maximum float64) []plot.Tick {
	if maximum <= minimum {
		return nil
	}

	ideal := (maximum - minimum) / 6
	magnitude := math.Pow(10, math.Floor(math.Log10(ideal)))
	step := 10 * magnitude

	for _, factor := range []float64{1, 2, 2.5, 5} {
		if ideal <= factor*magnitude*1.35 {
			step = factor * magnitude
			break
		}
	}

	ticks := make([]plot.Tick, 0, 8)

	for value := math.Ceil(minimum/step) * step; value <= maximum; value += step {
		label := strconv.FormatFloat(value, 'f', -1, 64)
		if math.Abs(value) < 1e8 {
			label = strconv.FormatFloat(value, 'f', 6, 64)
			label = strings.TrimRight(strings.TrimRight(label, "0"), ".")
		}

		ticks = append(ticks, plot.Tick{Value: value, Label: label})
	}

	return ticks
}
