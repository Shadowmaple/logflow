package field_getter

import (
	"regexp"
	"testing"
)

func TestReg(t *testing.T) {
	s := "kafka-%{[@metadata][kafka][topic]}-%{localtime}"
	r, _ := regexp.Compile(`%{(.+?)}`)

	// fmt.Println(r.FindAllString(s, -1))
	t.Logf("res: %v", r.FindAllStringSubmatch(s, -1))
	lastIdx := 0
	for _, v := range r.FindAllStringIndex(s, -1) {
		l, r := v[0], v[1]
		t.Logf("res: %v", s[lastIdx:l])
		t.Logf("res: %v", s[l+2:r-1])
		lastIdx = r
	}
	if lastIdx < len(s) {
		t.Logf("res: %v", s[lastIdx:])
	}
}
