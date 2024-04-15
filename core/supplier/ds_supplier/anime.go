package ds_supplier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/pkg/errors"
)

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
