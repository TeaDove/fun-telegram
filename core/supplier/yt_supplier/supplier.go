package yt_supplier

import (
	"context"
	"io"

	"github.com/kkdai/youtube/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Supplier struct {
	client youtube.Client
}

func New(ctx context.Context) (*Supplier, error) {
	client := youtube.Client{}

	r := Supplier{client: client}
	return &r, nil
}

func filterFormat(formats youtube.FormatList) youtube.Format {
	const (
		targetQualityLabel = "1080p"
	)

	withAudioChannels := formats.WithAudioChannels()
	if len(withAudioChannels) == 0 {
		return formats[0]
	}

	targetFormats := withAudioChannels.Quality(targetQualityLabel)
	if len(targetFormats) == 0 {
		return withAudioChannels[0]
	}

	return targetFormats[0]
}

func (r *Supplier) GetVideo(ctx context.Context, url string) (io.ReadCloser, error) {
	zerolog.Ctx(ctx).Info().Str("url", url).Msg("yt.video.requesting")
	video, err := r.client.GetVideoContext(ctx, url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get video from youtube")
	}

	format := filterFormat(video.Formats)
	stream, _, err := r.client.GetStream(video, &format)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get video stream")
	}

	zerolog.Ctx(ctx).Info().
		Str("quality", format.QualityLabel).
		Msg("yt.video.stream.connected")

	return stream, nil
}
