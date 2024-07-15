package ds_supplier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/tidwall/gjson"

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

type DSError struct {
	Detail string
	Code   int
}

func (r DSError) Error() string {
	return fmt.Sprintf("bad status code: %d, details: %s", r.Code, r.Detail)
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
		Str("elapsed", time.Since(t0).String()).
		Str("path", req.URL.Path).
		Msg("ds.request.done")

	if resp.StatusCode >= 400 {
		dsError := DSError{
			Detail: gjson.GetBytes(body, "detail").String(),
			Code:   resp.StatusCode,
		}

		return nil, errors.WithStack(dsError)
	}

	return body, nil
}
