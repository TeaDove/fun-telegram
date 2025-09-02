package container

import (
	"context"
	"fun_telegram/core/supplier/ds_supplier"
	"fun_telegram/core/supplier/gigachat_supplier"

	"github.com/pkg/errors"

	"fun_telegram/core/presentation/telegram"
	"fun_telegram/core/service/analitics"
)

type Container struct {
	Presentation *telegram.Presentation
}

func NewContainer(ctx context.Context) (Container, error) {
	dsSupplier := ds_supplier.New()

	analiticsService, err := analitics.New(dsSupplier)
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	protoClient, err := telegram.NewProtoClient(ctx)
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	gigachatSupplier, err := gigachat_supplier.NewSupplier(ctx)
	if err != nil {
		return Container{}, errors.WithStack(err)
	}

	telegramPresentation := telegram.MustNewTelegramPresentation(protoClient, analiticsService, gigachatSupplier)

	container := Container{telegramPresentation}

	return container, nil
}
