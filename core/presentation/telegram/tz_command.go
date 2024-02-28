package telegram

import (
	"fmt"
	"time"
)

// TODO move all chat settings to one place

const tzOffset int8 = 12

func int8ToLoc(tz int8) *time.Location {
	if tz == 0 {
		return time.FixedZone("UTC", 0)
	}

	if tz > 0 {
		return time.FixedZone(fmt.Sprintf("UTC+%d", tz), int(tz)*60*60)
	}

	return time.FixedZone(fmt.Sprintf("UTC%d", tz), int(tz)*60*60)
}
