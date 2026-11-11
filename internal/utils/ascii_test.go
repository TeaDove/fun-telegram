package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceNonASCIIWithSpace(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "sample", ReplaceNonASCIIWithSpace("sample\U0001FAE7🧸"))
}
