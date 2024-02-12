package analitics

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/wcharczuk/go-chart/v2"
	"image/jpeg"
	"image/png"
	"net/http"
)

func PngToJpeg(image []byte) ([]byte, error) {
	contentType := http.DetectContentType(image)

	switch contentType {
	case "image/png":
		// Decode the PNG image bytes
		img, err := png.Decode(bytes.NewReader(image))

		if err != nil {
			return nil, errors.WithStack(err)
		}

		buf := new(bytes.Buffer)

		if err = jpeg.Encode(buf, img, nil); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	return nil, errors.Errorf("unable to convert %#v to jpeg", contentType)
}

func getBarChart() chart.BarChart {
	return chart.BarChart{
		ColorPalette: chart.AlternateColorPalette,
		Width:        1000,
		Height:       1000,
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		BarWidth: 30,
		XAxis:    chart.Style{TextRotationDegrees: -90, FontSize: 13, TextHorizontalAlign: 7},
	}
}
