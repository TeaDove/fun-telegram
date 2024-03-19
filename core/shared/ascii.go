package shared

import (
	"strings"
	"unicode"
)

func ReplaceNonAsciiWithSpace(v string) string {
	return strings.TrimSpace(
		strings.Map(
			func(r rune) rune {
				if unicode.IsPrint(r) && !unicode.IsSymbol(r) {
					return r
				}

				return ' '
			},
			v,
		),
	)
}
