package ds_supplier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/teadove/teasutils/utils/closer_utils"

	"github.com/tidwall/gjson"

	"github.com/rs/zerolog"

	"fun_telegram/core/shared"

	"github.com/pkg/errors"
)

type Supplier struct {
	client   *http.Client
	basePath string
}

func New() *Supplier {
	r := Supplier{}
	r.client = http.DefaultClient
	r.basePath = shared.AppSettings.DsSupplierURL

	return &r
}

func (r *Supplier) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/health", r.basePath), nil)
	if err != nil {
		return errors.Wrap(err, "failed to make request")
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to do request")
	}

	closer_utils.CloseOrLog(ctx, resp.Body)

	if resp.StatusCode != http.StatusOK {
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
		http.MethodPost,
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

	closer_utils.CloseOrLog(ctx, resp.Body)

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
