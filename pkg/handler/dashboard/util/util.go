package util

import "time"

type util struct{}

func New() IUtil {
	return &util{}
}

func (u *util) TimeNow() time.Time {
	return time.Now()
}
