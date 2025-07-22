package shared

import (
	"time"

	"github.com/teadove/teasutils/utils/must_utils"
)

const (
	Undefined             = "undefined"
	Unknown               = "unknown"
	MaxUploadCount        = 500_000
	DefaultUploadCount    = 10_000
	MaxUploadQueryAge     = time.Hour * 24 * 365 * 2
	DefaultUploadQueryAge = time.Hour * 24 * 30 * 2

	TZInt = 3
	TZ    = "Europe/Moscow"
)

var TZTime = must_utils.Must(time.LoadLocation(TZ))
