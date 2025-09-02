package main

import (
	"fun_telegram/core/container"

	"github.com/teadove/teasutils/utils/logger_utils"
)

func main() {
	ctx := logger_utils.NewLoggedCtx()

	combatContainer, err := container.NewContainer(ctx)
	if err != nil {
		panic(err)
	}

	err = combatContainer.Presentation.Run(ctx)
	if err != nil {
		panic(err)
	}
}
