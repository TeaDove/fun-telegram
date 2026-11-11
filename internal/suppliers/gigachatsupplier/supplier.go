package gigachatsupplier

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/TeaDove/fun-telegram/internal/utils"
	"github.com/pkg/errors"
)

type Supplier struct {
	accessToken string
	settings    utils.GigachatSettings

	httpClient *http.Client
}

func NewSupplier(ctx context.Context) (*Supplier, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint: gosec // FIXME
		},
	}

	return newSupplier(ctx, utils.AppSettings.Gigachat, client)
}

func newSupplier(ctx context.Context, settings utils.GigachatSettings, client *http.Client) (*Supplier, error) {
	r := &Supplier{httpClient: client, settings: settings}

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

	resp, err := r.sendRequestWithAuth(ctx, body)
	if err != nil {
		return "", errors.Wrap(err, "failed to do request")
	}

	defer utils.CloseOrLog(ctx, resp.Body)

	var response Response

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", errors.WithStack(err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("no choices found")
	}

	return response.Choices[0].Message.Content, nil
}
