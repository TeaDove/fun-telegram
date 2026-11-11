package container

import (
	"context"

	"github.com/TeaDove/fun-telegram/internal/suppliers/dssupplier"
	"github.com/TeaDove/fun-telegram/internal/suppliers/gigachatsupplier"

	"github.com/pkg/errors"

	"github.com/TeaDove/fun-telegram/internal/apis/telegramapi"
	"github.com/TeaDove/fun-telegram/internal/services/analyticsservice"
)

type Container struct {
	Presentation *telegramapi.Presentation
}

func NewContainer(ctx context.Context) (Container, error) {
	dsSupplier := dssupplier.New()

	analiticsService, err := analyticsservice.New(dsSupplier)
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	protoClient, err := telegramapi.NewProtoClient(ctx)
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	gigachatSupplier, err := gigachatsupplier.NewSupplier(ctx)
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	telegramPresentation := telegramapi.MustNewTelegramPresentation(protoClient, analiticsService, gigachatSupplier)

	container := Container{telegramPresentation}

	return container, nil
}
