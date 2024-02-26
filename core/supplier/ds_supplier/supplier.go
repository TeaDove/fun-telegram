package ds_supplier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/teadove/fun_telegram/core/shared"
)

type Supplier struct {
	client   *http.Client
	basePath string
}

func New(ctx context.Context) (*Supplier, error) {
	r := Supplier{}
	r.client = http.DefaultClient
	r.basePath = shared.AppSettings.DsSupplierUrl

	return &r, nil
}

func (r *Supplier) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/health", r.basePath), nil)
	if err != nil {
		return errors.Wrap(err, "failed to make request")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to do request")
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("wrong status code: %d", resp.StatusCode)
	}

	return nil
}

func (r *Supplier) sendRequest(ctx context.Context, path string, input any) ([]byte, error) {
	reqBody, err := json.Marshal(&input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/%s", r.basePath, path), bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}

	t0 := time.Now()

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	zerolog.Ctx(ctx).Debug().Str("status", "ds.request.done").Dur("elapsed", time.Since(t0)).Str("path", path).Send()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	if resp.StatusCode >= 400 {
		return nil, errors.Errorf("bad status code: %d, content: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

type DrawInput struct {
	Title  string `json:"title,omitempty"`
	XLabel string `json:"xlabel,omitempty"`
	YLabel string `json:"ylabel,omitempty"`
}

type DrawBarInput struct {
	DrawInput

	Values map[string]float64 `json:"values"`
	Limit  int                `json:"limit,omitempty"`
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

type DrawGraphInput struct {
	DrawInput

	Edges []GraphEdge `json:"edges,omitempty"`
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
