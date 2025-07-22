package main

import (
	"github.com/teadove/fun_telegram/core/container"
	"github.com/teadove/teasutils/utils/logger_utils"
)

func main() {
	ctx := logger_utils.NewLoggedCtx()
	combatContainer := container.MustNewCombatContainer(ctx)

	err := combatContainer.Presentation.Run(ctx)
	if err != nil {
		panic(err)
	}
}
