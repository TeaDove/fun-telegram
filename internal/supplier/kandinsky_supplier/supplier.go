package kandinsky_supplier

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

type Supplier struct {
	client http.Client
	url    string
	key    string
	secret string
	model  int
}

func New(ctx context.Context, key, secret string) (*Supplier, error) {
	r := Supplier{
		client: http.Client{},
		url:    "https://api-key.fusionbrain.ai",
		key:    key,
		secret: secret,
	}

	model, err := r.getModels(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.model = model

	return &r, nil
}

func (r *Supplier) addCredsHeaders(req *http.Request) {
	req.Header.Add("X-Key", fmt.Sprintf("Key %s", r.key))
	req.Header.Add("X-Secret", fmt.Sprintf("Secret %s", r.secret))
}
