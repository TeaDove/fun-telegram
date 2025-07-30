package telegram

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_GetArguments_SilentArgument_Ok(t *testing.T) {
	t.Parallel()

	input := getOpt("!ping --silent")

	assert.True(t, input.Silent)

	input = getOpt("!ping")

	assert.False(t, input.Silent)
}

func TestUnit_GetArguments_TextCompiled_Ok(t *testing.T) {
	t.Parallel()

	input := getOpt("!ping --silent user not found hi!")

	assert.True(t, input.Silent)
	assert.Equal(t, "user not found hi!", input.Text)
}

func TestUnit_GetArguments_Arguments_Ok(t *testing.T) {
	t.Parallel()

	input := getOpt(`!ping -q --negative="bad input" user not found hi!`, optFlag{Long: "negative"})

	assert.True(t, input.Silent)
	assert.Equal(t, "bad input", input.Ops["negative"])
	assert.Equal(t, "user not found hi!", input.Text)
}

func TestUnit_StripWords_WithSpace_Ok(t *testing.T) {
	t.Parallel()

	words := stripWords("one two three")

	assert.Equal(t, []string{"one", "two", "three"}, words)
}

func TestUnit_StripWords_None_Ok(t *testing.T) {
	t.Parallel()

	words := stripWords("")

	assert.Equal(t, []string{}, words)
}

func TestUnit_StripWords_WithQuotes_Ok(t *testing.T) {
	t.Parallel()

	words := stripWords(`one "two three" four`)

	assert.Len(t, words, 3)
	assert.Equal(t, []string{"one", "two three", "four"}, words)
}

func TestUnit_StripWords_QuotesNotFromStart_Ok(t *testing.T) {
	t.Parallel()

	words := stripWords(`!ping --silent --negative="bad input" user not found hi!`)

	assert.Len(t, words, 7)
	assert.Equal(
		t,
		[]string{"!ping", "--silent", `--negative=bad input`, "user", "not", "found", "hi!"},
		words,
	)
}

func TestUnit_GetArguments_LongDash_Ok(t *testing.T) {
	t.Parallel()

	input := getOpt("!ping —silent")

	assert.True(t, input.Silent)
}
