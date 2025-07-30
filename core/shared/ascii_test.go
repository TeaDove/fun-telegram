package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_Shared_TrimUnprintable_Ok(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "nastik  ", ReplaceNonASCIIWithSpace("nastik\U0001FAE7🧸"))
}
