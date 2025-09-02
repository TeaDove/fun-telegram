package gigachat_supplier

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fun_telegram/core/shared"
	"net/http"

	"github.com/pkg/errors"
	"github.com/teadove/teasutils/utils/closer_utils"
)

type Supplier struct {
	accessToken string

	httpClient *http.Client
}

func NewSupplier(ctx context.Context) (*Supplier, error) {
	r := &Supplier{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint: gosec // FIXME
			},
		},
	}

	err := r.auth(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to auth")
	}

	return r, nil
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Completion struct {
	Model             string    `json:"model"`
	Messages          []Message `json:"messages"`
	Stream            bool      `json:"stream"`
	RepetitionPenalty int       `json:"repetition_penalty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type Response struct {
	Choices []Choice `json:"choices"`
}

func (r *Supplier) OneMessage(ctx context.Context, messages []Message) (string, error) {
	body := Completion{Model: "GigaChat", Messages: messages, Stream: false, RepetitionPenalty: 1}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return "", errors.WithStack(err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		shared.AppSettings.Gigachat.BaseURL,
		bytes.NewReader(bodyJSON),
	)
	if err != nil {
		return "", errors.WithStack(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.sendRequestWithAuth(ctx, req)
	if err != nil {
		return "", errors.Wrap(err, "failed to do request")
	}

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", errors.WithStack(err)
	}

	closer_utils.CloseOrLog(ctx, resp.Body)

	if len(response.Choices) == 0 {
		return "", errors.New("no choices found")
	}

	return response.Choices[0].Message.Content, nil
}
