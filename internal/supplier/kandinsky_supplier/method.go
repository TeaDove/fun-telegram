package kandinsky_supplier

import (
	"bytes"
	"context"
	"encoding/json"
	goErrors "errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tidwall/gjson"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrImageWasCensored    = errors.New("image was censored")
	ErrImageNotReady       = errors.New("image not ready yet")
	ErrImageCreationFailed = errors.New("image cannot be created")
)

func (r *Supplier) getModels(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/key/api/v1/models", r.url), nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	r.addCredsHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	if resp.StatusCode != 200 {
		return 0, errors.Errorf("bad status code, status code: %d, content: %s", resp.StatusCode, string(respBytes))
	}

	for _, v := range gjson.ParseBytes(respBytes).Array() {
		model := int(v.Get("id").Int())
		zerolog.Ctx(ctx).Info().Str("status", "kandinsky.model.got").Int("model_id", model).Send()

		return model, nil
	}

	return 0, errors.Errorf("failed to parse model from response, content: %s", string(respBytes))
}

type RequestGenerationInput struct {
	Prompt               string
	NegativePromptUnclip string
	Style                string
}

type RequestGenerationRequest struct {
	Type                 string `json:"type"`
	Width                int    `json:"width"`
	Height               int    `json:"height"`
	Style                string `json:"style,omitempty"`
	NegativePromptUnclip string `json:"negativePromptUnclip"`
	GenerateParams       struct {
		Query string `json:"query"`
	} `json:"generateParams"`
}

func (r *Supplier) RequestGeneration(ctx context.Context, input *RequestGenerationInput) (uuid.UUID, error) {
	zerolog.Ctx(ctx).Info().Str("status", "kandinsky.image.generation.begin").Interface("input", input).Send()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	paramsPart := make(map[string][]string)
	paramsPart["Content-Disposition"] = append(paramsPart["Content-Disposition"], "form-data; name=\"params\"")
	paramsPart["Content-Type"] = append(paramsPart["Content-Type"], "application/json")

	paramsWriter, err := writer.CreatePart(paramsPart)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	paramsPayload := RequestGenerationRequest{
		Type:                 "GENERATE",
		Width:                1024,
		Height:               1024,
		Style:                input.Style,
		NegativePromptUnclip: input.NegativePromptUnclip,
		GenerateParams: struct {
			Query string `json:"query"`
		}{Query: input.Prompt},
	}

	paramsPayloadBytes, err := json.Marshal(&paramsPayload)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	_, err = paramsWriter.Write(paramsPayloadBytes)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	err = writer.WriteField("model_id", strconv.Itoa(r.model))
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	err = writer.Close()
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/key/api/v1/text2image/run", r.url), payload)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.addCredsHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}

	imageId, err := uuid.Parse(gjson.GetBytes(respBytes, "uuid").String())
	if err != nil {
		zerolog.Ctx(ctx).Warn().Str("status", "kandinsky.image.cannot.be.generated").Send()
		return uuid.Nil, goErrors.Join(ErrImageCreationFailed, err)
	}

	zerolog.Ctx(ctx).Info().Str("status", "kandinsky.image.generation.send").Send()

	return imageId, nil
}

func (r *Supplier) Get(ctx context.Context, id uuid.UUID) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("%s/key/api/v1/text2image/status/%s", r.url, id.String()),
		nil,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r.addCredsHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if gjson.GetBytes(respBytes, "status").String() == "DONE" {
		for _, img := range gjson.GetBytes(respBytes, "images").Array() {
			zerolog.Ctx(ctx).Info().Str("status", "kandinsky.image.generation.done").Str("id", id.String()).Send()
			return []byte(img.String()), nil
		}

		return nil, errors.WithStack(ErrImageCreationFailed)
	}

	if gjson.GetBytes(respBytes, "censored").Bool() {
		zerolog.Ctx(ctx).Info().Str("status", "kandinsky.image.censored").Str("id", id.String()).Send()
		return nil, errors.WithStack(ErrImageWasCensored)
	}

	zerolog.Ctx(ctx).Info().Str("status", "kandinsky.image.not.ready").Str("id", id.String()).Send()

	return nil, errors.WithStack(ErrImageNotReady)
}

const delay = time.Second * 8
const maxAttempts = 5

func (r *Supplier) WaitGet(ctx context.Context, id uuid.UUID) ([]byte, error) {
	amount := maxAttempts

	for amount > 0 {
		img, err := r.Get(ctx, id)
		if err != nil {
			if errors.Is(err, ErrImageNotReady) {
				time.Sleep(delay)
				amount--

				continue
			}

			return nil, errors.WithStack(err)
		}

		return img, nil
	}

	zerolog.Ctx(ctx).Info().Str("status", "kandinsky.image.creation.failed").Str("id", id.String()).Send()

	return nil, errors.WithStack(ErrImageCreationFailed)
}

func (r *Supplier) WaitGeneration(ctx context.Context, input *RequestGenerationInput) ([]byte, error) {
	id, err := r.RequestGeneration(ctx, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	time.Sleep(delay)

	img, err := r.WaitGet(ctx, id)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return img, nil
}
