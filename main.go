package main

import (
	"github.com/TeaDove/fun-telegram/internal/container"
	"github.com/TeaDove/fun-telegram/internal/utils"

	"github.com/teadove/teasutils/serviceutils/loggerutils"
)

func main() {
	ctx := loggerutils.NewLoggedCtx()

	settings, err := utils.LoadSettings()
	if err != nil {
		panic(err)
	}

	combatContainer, err := container.NewContainer(ctx, settings.Telegram)
	if err != nil {
		panic(err)
	}

	err = combatContainer.Presentation.Run(ctx)
	if err != nil {
		panic(err)
	}
}
