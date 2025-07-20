package shared

import "time"

const (
	Undefined             = "undefined"
	Unknown               = "unknown"
	MaxUploadCount        = 500_000
	DefaultUploadCount    = 10_000
	MaxUploadQueryAge     = time.Hour * 24 * 365 * 2
	DefaultUploadQueryAge = time.Hour * 24 * 30 * 2
	TZ                    = "Europe/Moscow"
)
