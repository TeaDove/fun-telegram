package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_Shared_TrimUnprintable_Ok(t *testing.T) {
	assert.Equal(t, "nastik  ", ReplaceNonAsciiWithSpace("nastik\U0001FAE7🧸"))
}
