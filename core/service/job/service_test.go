package job

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teadove/fun_telegram/core/shared"
)

func TestIntegration_JobService_DeleteMessage_Ok(t *testing.T) {
	r := getService(t)

	_, err := r.DeleteOldMessages(shared.GetCtx())
	assert.NoError(t, err)
}
