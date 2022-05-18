package util

import (
	"time"
)

const (
	TimeFormat     = "2006-01-02 15:04:05"
	GormTimeFormat = "2006-01-02T15:04:05Z"
)

func TimeToString(time time.Time) string {
	return time.Format(TimeFormat)
}

func StringToTime(str string) time.Time {
	t, _ := time.ParseInLocation(GormTimeFormat, str, time.Local)
	return t
}
