package field

import (
	"regexp"

	"github.com/Shadowmaple/logflow/internal/event"
)

// 用于对某个字段设置值
type FieldSetter interface {
	SetField(event *event.Event, value any)
}

// NewFieldSetter creates a new FieldSetter.
// It returns OneLevelFieldSetter if [xxx]/xxx passed
// It returns MultiLevelFieldSetter if [xxx][yyy] passed
// TODO: jsonpath
func NewFieldSetter(template string, overwrite bool) FieldSetter {
	matchp, _ := regexp.Compile(`(\[.*?\])+`)
	findp, _ := regexp.Compile(`(\[(.*?)\])`)
	if matchp.Match([]byte(template)) {
		fields := make([]string, 0)
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}
		if len(fields) == 1 {
			return newOneLevelFieldSetter(fields[0], overwrite)
		}
		return newMultiLevelFieldSetter(fields, overwrite)
	}
	return newOneLevelFieldSetter(template, overwrite)
}
