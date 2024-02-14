package analitics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_ToxicityService_GetToxicity_Ok(t *testing.T) {
	r := getService(t)

	isToxicWord, err := r.IsToxic("заеб!")
	assert.NoError(t, err)
	assert.True(t, isToxicWord)

	isToxicWord, err = r.IsToxic("привет")
	assert.NoError(t, err)
	assert.False(t, isToxicWord)

	isToxicWord, err = r.IsToxic("ты меня заебал")
	assert.NoError(t, err)
	assert.True(t, isToxicWord)
}
