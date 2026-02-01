package utils

import "time"

func FormatTime(t time.Time) string {
	return t.Format(time.DateTime)
}

func ValidateKeyExists(m map[string]any, keys []string) bool {
	return true
}
