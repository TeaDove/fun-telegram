package main

import (
	"github.com/teadove/goteleout/internal/container"
)

func main() {
	combatContainer := container.MustNewCombatContainer()
	combatContainer.ClientService.Run()
}
