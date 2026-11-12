package main

import (
	"os"

	"github.com/TeaDove/fun-telegram/internal/apis/telegramapi"
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

	if len(os.Args) == 2 && os.Args[1] == "login-qr" {
		err := telegramapi.LoginQR(ctx, settings.Telegram, os.Stdin, os.Stdout)
		if err != nil {
			panic(err)
		}

		return
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
