package main

import (
	"github.com/TeaDove/fun-telegram/internal/container"

	"github.com/teadove/teasutils/serviceutils/loggerutils"
)

func main() {
	ctx := loggerutils.NewLoggedCtx()

	combatContainer, err := container.NewContainer(ctx)
	if err != nil {
		panic(err)
	}

	err = combatContainer.Presentation.Run(ctx)
	if err != nil {
		panic(err)
	}
}
