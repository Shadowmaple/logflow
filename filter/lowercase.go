package filter

import (
	"errors"
	"strings"

	"github.com/Shadowmaple/logflow/field/field_setter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/model"
	"github.com/Shadowmaple/logflow/render"
)

type LowercaseFilter struct {
	fields map[field_setter.FieldSetter]render.Render
}

func init() {
	register("lowercase", newLowercaseFilter)
}

func newLowercaseFilter(config map[string]any) model.Filter {
	fields, ok := config["fields"].([]string)
	if !ok {
		panic("lowercase invalid fields")
	}
	mp := make(map[field_setter.FieldSetter]render.Render)
	for _, fieldName := range fields {
		mp[field_setter.NewFieldSetter(fieldName, true)] = render.GetRender(fieldName)
	}
	return &LowercaseFilter{fields: mp}
}

// 若遇到一个字段处理失败，则继续处理其它的，最后返回错误
func (f *LowercaseFilter) Filter(event *event.Event) (*event.Event, error) {
	var failed bool
	for fieldSetter, render := range f.fields {
		if value, err := render.Render(event); err == nil {
			v, ok := value.(string)
			if !ok {
				failed = true
				continue
			}
			fieldSetter.SetField(event, strings.ToLower(v))
		} else {
			failed = true
		}
	}
	if failed {
		return event, errors.New("lowercase filter failed")
	}
	return event, nil
}
