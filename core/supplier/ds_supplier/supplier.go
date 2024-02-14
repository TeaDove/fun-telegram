package ds_supplier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/teadove/goteleout/core/shared"
	"io"
	"net/http"
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

type DrawBarInput struct {
	Values map[string]float64 `json:"values"`
	Title  string             `json:"title"`
	XLabel string             `json:"xlabel"`
	YLabel string             `json:"ylabel"`
	Limit  int                `json:"limit,omitempty"`
}

func (r *Supplier) DrawBar(ctx context.Context, input *DrawBarInput) ([]byte, error) {
	reqBody, err := json.Marshal(&input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/histogram", r.basePath), bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	if resp.StatusCode >= 400 {
		return nil, errors.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	return body, nil
}
