package analitics

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnit_ToxicityService_GetToxicity_Ok(t *testing.T) {
	r := getService(t)

	isToxicWord, err := r.IsToxicWord("заеб!")
	assert.NoError(t, err)
	assert.True(t, isToxicWord)

	isToxicWord, err = r.IsToxicWord("привет")
	assert.NoError(t, err)
	assert.False(t, isToxicWord)
}
