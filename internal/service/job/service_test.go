package job

import (
	"testing"
)

func TestIntegration_JobService_DeleteMessage_Ok(t *testing.T) {
	r := getService(t)

	r.DeleteOldMessages()
}
