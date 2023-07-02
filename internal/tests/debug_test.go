package tests

import (
	"github.com/teadove/goteleout/internal/container"
	"testing"
)

var combatContainer = container.MustNewCombatContainer()

func TestIntegration_ClientService_Run_Ok(t *testing.T) {
	combatContainer.ClientService.Run()
}
