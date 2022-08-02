package utils

import "time"

const timeFormatStr = "2006-01-02 15:04:05"

func TimeFormat(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(timeFormatStr)
}
