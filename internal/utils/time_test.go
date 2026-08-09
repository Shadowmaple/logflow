package utils

import (
	"testing"
)

func TestConvertToGoFormat(t *testing.T) {
	format := "yyyy-MM-dd HH:mm:ss.SSS"
	expected := "2006-01-02 15:04:05.000"
	goFormat := ConvertToGoFormat(format)
	t.Log(goFormat)
	if goFormat != expected {
		t.Errorf("ConvertToGoFormat(%s) = %s; want %s", format, goFormat, expected)
	}
}
