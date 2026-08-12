package utils

import (
	"testing"
)

func TestConvertToGoFormat(t *testing.T) {
	formats := []string{
		"yyyy-MM-dd HH:mm:ss.SSS",
		"yyyy-MM-dd HH:mm:ss",
		"yyyy-MM-dd",
		"HH:mm:ss",
		"yyyy-MM-dd HH:mm",
		"yyyy/MM/dd HH:mm:ss.SSS",
	}
	expectedList := []string{
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05.000",
	}
	for i, format := range formats {
		goFormat := ConvertToGoFormat(format)
		// if err != nil {
		// 	t.Errorf("ConvertToGoFormat(%s) = %v", format, err)
		// 	continue
		// }
		t.Log(goFormat)
		if goFormat != expectedList[i] {
			t.Errorf("ConvertToGoFormat(%s) = %s; want %s", format, goFormat, expectedList[i])
		}
	}
}
