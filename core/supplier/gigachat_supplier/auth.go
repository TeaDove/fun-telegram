package gigachat_supplier

import (
	"context"
	"fun_telegram/core/shared"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/teadove/teasutils/utils/closer_utils"
	"github.com/teadove/teasutils/utils/test_utils"
	"github.com/tidwall/gjson"
)

func (r *Supplier) auth(ctx context.Context) error {
	body := strings.NewReader(`scope=GIGACHAT_API_PERS`)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, shared.AppSettings.Gigachat.AuthURL, body)
	if err != nil {
		return errors.WithStack(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("RqUID", uuid.New().String())
	req.Header.Set("Authorization", "Basic "+shared.AppSettings.Gigachat.AuthorizationKey)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to do request")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WithStack(err)
	}

	closer_utils.CloseOrLog(ctx, resp.Body)

	accessToken := gjson.GetBytes(bodyBytes, "access_token").String()
	if accessToken == "" {
		return errors.New("access token is empty")
	}

	r.accessToken = accessToken

	zerolog.Ctx(ctx).Info().Msg("gigachat.autorized")

	return nil
}

func (r *Supplier) sendRequestWithAuth(ctx context.Context, req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+r.accessToken)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode != http.StatusUnauthorized {
			return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		err = r.auth(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to autorize")
		}

		req.Header.Set("Authorization", "Bearer "+r.accessToken)

		resp, err = r.httpClient.Do(req)
		if err != nil {
			return nil, errors.Wrap(err, "failed to do request")
		}

		if resp.StatusCode != http.StatusOK {
			cont, _ := io.ReadAll(resp.Body)
			test_utils.Pprint(string(cont))

			return nil, errors.Errorf("unexpected status code after reauth: %d", resp.StatusCode)
		}

		return resp, nil
	}

	return resp, nil
}
