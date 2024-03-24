package ds_supplier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"

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

	shared.CloseOrLog(ctx, resp.Body)

	if resp.StatusCode != 200 {
		return errors.Errorf("wrong status code: %d", resp.StatusCode)
	}

	return nil
}

func (r *Supplier) sendRequest(ctx context.Context, path string, input any) ([]byte, error) {
	//zerolog.Ctx(ctx).
	//	Debug().
	//	Str("status", "ds.request.sending").
	//	Interface("input", input).
	//	Str("path", path).
	//	Send()

	reqBody, err := json.Marshal(&input)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal request body")
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("%s/%s", r.basePath, path),
		bytes.NewReader(reqBody),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}

	resp, err := r.doRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	return resp, nil
}

func (r *Supplier) doRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	t0 := time.Now()

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read body request")
	}

	shared.CloseOrLog(ctx, resp.Body)

	zerolog.Ctx(ctx).Debug().
		Str("status", "ds.request.done").
		Dur("elapsed", time.Since(t0)).
		Str("path", req.URL.Path).Send()

	if resp.StatusCode >= 400 {
		return nil, errors.Errorf("bad status code: %d, content: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
