package shared

import (
	"math/rand/v2"
	"strings"
)

func RandomString() string {
	const alfabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var builder strings.Builder
	for i := 0; i < 50; i++ {
		builder.WriteByte(alfabet[rand.IntN(len(alfabet))])
	}

	return builder.String()
}
