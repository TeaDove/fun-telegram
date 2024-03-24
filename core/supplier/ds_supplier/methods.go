package ds_supplier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"mime/multipart"
	"net/http"
	"time"
)

type DrawInput struct {
	Title       string `json:"title,omitempty"`
	XLabel      string `json:"xlabel,omitempty"`
	YLabel      string `json:"ylabel,omitempty"`
	FigSize     []int  `json:"figsize,omitempty"`
	ImageFormat string `json:"image_format,omitempty"`
}

type DrawBarInput struct {
	DrawInput

	Values map[string]float64 `json:"values"`
	Limit  int                `json:"limit,omitempty"`
	Asc    bool               `json:"asc"`
}

func (r *Supplier) DrawBar(ctx context.Context, input *DrawBarInput) ([]byte, error) {
	body, err := r.sendRequest(ctx, "histogram", input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to draw")
	}

	return body, nil
}

type DrawTimeseriesInput struct {
	DrawInput

	Values   map[string]map[time.Time]float64 `json:"values"`
	OnlyTime bool                             `json:"only_time"`
}

func (r *Supplier) DrawTimeseries(ctx context.Context, input *DrawTimeseriesInput) ([]byte, error) {
	body, err := r.sendRequest(ctx, "timeseries", input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to draw")
	}

	return body, nil
}

type GraphEdge struct {
	First  string  `json:"first"`
	Second string  `json:"second"`
	Weight float64 `json:"weight"`
}

type GraphNode struct {
	Image  []byte  `json:"image,omitempty"`
	Weight float64 `json:"weight,omitempty"`
}

type DrawGraphInput struct {
	DrawInput

	Edges         []GraphEdge          `json:"edges,omitempty"`
	Layout        string               `json:"layout,omitempty"`
	WeightedEdges bool                 `json:"weighted_edges"`
	Nodes         map[string]GraphNode `json:"nodes,omitempty"`
	RootNode      string               `json:"root_node,omitempty"`
}

func (r *Supplier) DrawGraph(ctx context.Context, input *DrawGraphInput) ([]byte, error) {
	body, err := r.sendRequest(ctx, "graph", input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to draw")
	}

	return body, nil
}

func (r *Supplier) DrawGraphAsHeatpmap(ctx context.Context, input *DrawGraphInput) ([]byte, error) {
	body, err := r.sendRequest(ctx, "graph-as-heatmap", input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to draw")
	}

	return body, nil
}

func (r *Supplier) AnimePrediction(ctx context.Context, image []byte) (float64, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part1, err := writer.CreateFormFile("image", "image.webp")
	if err != nil {
		return 0, errors.Wrap(err, "failed to create form file")
	}

	_, err = part1.Write(image)
	if err != nil {
		return 0, errors.Wrap(err, "failed to write image")
	}

	err = writer.Close()
	if err != nil {
		return 0, errors.Wrap(err, "failed to close writer")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/%s", r.basePath, "anime/predict"),
		payload,
	)
	if err != nil {
		return 0, errors.Wrap(err, "failed to make request")
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := r.doRequest(ctx, req)
	if err != nil {
		return 0, errors.Wrap(err, "failed to do request")
	}

	response := struct {
		Prediction float64 `json:"prediction"`
	}{}

	err = json.Unmarshal(resp, &response)
	if err != nil {
		return 0, errors.Wrap(err, "failed to unmarshal response")
	}

	return response.Prediction, nil
}
