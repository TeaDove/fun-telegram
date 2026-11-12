package container

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog"

	"github.com/TeaDove/fun-telegram/internal/apis/telegramapi"
	"github.com/TeaDove/fun-telegram/internal/services/analyticsservice"
	"github.com/TeaDove/fun-telegram/internal/utils"
)

type Container struct {
	Presentation *telegramapi.Presentation
}

func NewContainer(ctx context.Context, settings utils.TelegramSettings) (Container, error) {
	analiticsService, err := analyticsservice.New()
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	protoClient, err := telegramapi.NewProtoClient(ctx, settings)
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	telegramPresentation := telegramapi.MustNewTelegramPresentation(protoClient, analiticsService)

	container := Container{telegramPresentation}

	zerolog.Ctx(ctx).Info().Msg("container.built")

	return container, nil
}
