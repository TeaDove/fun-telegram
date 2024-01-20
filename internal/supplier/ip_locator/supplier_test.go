package ip_locator

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/teadove/goteleout/internal/utils"
	"testing"
)

func TestIntegration_IpLocator_GetLocation_Ok(t *testing.T) {
	r := Supplier{}

	resp, err := r.GetLocation(context.Background(), "116.203.245.151")

	assert.NoError(t, err)
	utils.SendInterface(resp)
}

func TestIntegration_IpLocator_GetLocation_ValidationError(t *testing.T) {
	r := Supplier{}

	_, err := r.GetLocation(context.Background(), "116.203.299.151")

	assert.Error(t, err)
	assert.Equal(t, "request failed with error: invalid query", err.Error())
}
