package filter

import (
	"errors"
	"strings"

	"github.com/Shadowmaple/logflow/field/field_getter"
	"github.com/Shadowmaple/logflow/field/field_setter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/model"
)

type UppercaseFilter struct {
	fields map[field_setter.FieldSetter]field_getter.FieldGetter
}

func init() {
	register("uppercase", newUppercaseFilter)
}

func newUppercaseFilter(config map[string]any) model.Filter {
	fields, ok := config["fields"].([]string)
	if !ok {
		panic("uppercase invalid fields")
	}
	mp := make(map[field_setter.FieldSetter]field_getter.FieldGetter)
	for _, f := range fields {
		mp[field_setter.NewFieldSetter(f, true)] = field_getter.GetFieldGetter(f)
	}
	return &UppercaseFilter{fields: mp}
}

// 若遇到一个字段处理失败，则继续处理其它的，最后返回错误
func (f *UppercaseFilter) Filter(event *event.Event) (*event.Event, error) {
	var failed bool
	for fieldSetter, getter := range f.fields {
		if value, err := getter.GetField(event); err == nil {
			v, ok := value.(string)
			if !ok {
				failed = true
				continue
			}
			fieldSetter.SetField(event, strings.ToUpper(v))
		} else {
			failed = true
		}
	}
	if failed {
		return event, errors.New("uppercase filter failed")
	}
	return event, nil
}
