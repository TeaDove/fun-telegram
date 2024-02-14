package job

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teadove/goteleout/core/utils"
)

func TestIntegration_JobService_DeleteMessage_Ok(t *testing.T) {
	r := getService(t)

	_, err := r.DeleteOldMessages(utils.GetCtx())
	assert.NoError(t, err)
}
