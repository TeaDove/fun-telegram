package container

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog"

	"github.com/TeaDove/fun-telegram/internal/apis/telegramapi"
	"github.com/TeaDove/fun-telegram/internal/services/analyticsservice"
)

type Container struct {
	Presentation *telegramapi.Presentation
}

func NewContainer(ctx context.Context) (Container, error) {
	analiticsService, err := analyticsservice.New()
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	protoClient, err := telegramapi.NewProtoClient(ctx)
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	telegramPresentation := telegramapi.MustNewTelegramPresentation(protoClient, analiticsService)

	container := Container{telegramPresentation}

	zerolog.Ctx(ctx).Info().Msg("container.built")

	return container, nil
}
