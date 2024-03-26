package shared

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(text string) string {
	hasher := sha256.New()
	hasher.Write([]byte(text))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return sha
}
