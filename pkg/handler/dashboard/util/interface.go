package util

import "time"

type IUtil interface {
	TimeNow() time.Time
}
