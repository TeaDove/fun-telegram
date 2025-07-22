package telegram

import (
	"strings"
	"time"
)

type optFlag struct {
	Short       string
	Long        string
	Description string
}

type input struct {
	Text      string
	Silent    bool
	Ops       map[string]string
	StartedAt time.Time
}

var FlagSilent = optFlag{Long: "silent", Short: "q"}

func stripWords(text string) []string {
	output := make([]string, 0, 10)

	var (
		start    int
		idx      int
		quoted   bool
		quoteIdx int
		buff     strings.Builder
	)

	for idx < len(text) {
		if text[idx] == '"' {
			if quoted {
				quoted = false

				buff.Reset()
				buff.WriteString(text[start:quoteIdx] + text[quoteIdx+1:idx])

				idx += 1

				continue
			} else {
				quoteIdx = idx
				quoted = true
			}
		}

		if text[idx] == ' ' && !quoted {
			buffString := buff.String()
			if buffString != "" {
				output = append(output, buffString)
			}

			buff.Reset()

			start = idx + 1
			idx += 1

			continue
		}

		buff.WriteByte(text[idx])
		idx += 1
	}

	if start != idx {
		output = append(output, text[start:idx])
	}

	return output
}

// GetOpt
// nolint: gocyclo
func GetOpt(text string, flags ...optFlag) (output input) {
	const longHypenByte = 226

	flags = append(flags, FlagSilent)
	longs := make(map[string]struct{}, 4)
	shortToLong := make(map[string]string, 4)

	for _, flag := range flags {
		longs[flag.Long] = struct{}{}
		shortToLong[flag.Short] = flag.Long
	}

	output.Ops = make(map[string]string, 3)
	textBuilder := strings.Builder{}
	words := stripWords(text)

	if len(words) <= 1 {
		return output
	}

	for _, word := range words[1:] {
		if len(word) == 0 || len(word) == 1 || !(word[0] == '-' || word[0] == longHypenByte) {
			textBuilder.WriteString(" " + word)
			continue
		}

		var (
			long      string
			statement string
		)

		splited := strings.Split(word, "=")

		switch {
		case word[0] == longHypenByte:
			if len(word) == 1 {
				textBuilder.WriteString(" " + word)
				continue
			}

			long = splited[0][3:]

			if len(splited) > 1 {
				statement = splited[1]
			}
		case word[0:2] == "--":
			if len(word) == 2 {
				textBuilder.WriteString(" " + word)
				continue
			}

			long = splited[0][2:]

			if len(splited) > 1 {
				statement = splited[1]
			}
		default:
			var ok bool

			long, ok = shortToLong[splited[0][1:]]
			if !ok {
				textBuilder.WriteString(" " + word)
				continue
			}

			if len(splited) > 1 {
				statement = splited[1]
			}
		}

		_, ok := longs[long]
		if !ok {
			textBuilder.WriteString(" " + word)
			continue
		}

		output.Ops[long] = strings.Trim(statement, `"`)
	}

	_, ok := output.Ops[FlagSilent.Long]
	if ok {
		output.Silent = true
	}

	delete(output.Ops, FlagSilent.Long)

	output.Text = strings.Trim(textBuilder.String(), " ")

	return output
}
