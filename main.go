package main

import (
	"os"

	"github.com/TeaDove/fun-telegram/internal/apis/telegramapi"
	"github.com/TeaDove/fun-telegram/internal/container"

	"github.com/teadove/teasutils/serviceutils/loggerutils"
)

func main() {
	ctx := loggerutils.NewLoggedCtx()

	if len(os.Args) == 2 && os.Args[1] == "login-qr" {
		err := telegramapi.LoginQR(ctx, os.Stdin, os.Stdout)
		if err != nil {
			panic(err)
		}

		return
	}

	combatContainer, err := container.NewContainer(ctx)
	if err != nil {
		panic(err)
	}

	err = combatContainer.Presentation.Run(ctx)
	if err != nil {
		panic(err)
	}
}
