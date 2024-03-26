package shared

import (
	"math/rand/v2"
	"strings"
)

func RandomString() string {
	return RandomStringWithLength(10)
}

func RandomStringWithLength(length int) string {
	const alfabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	if length <= 0 {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < length; i++ {
		builder.WriteByte(alfabet[rand.IntN(len(alfabet))])
	}

	return builder.String()
}
