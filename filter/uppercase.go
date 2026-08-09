package filter

import (
	"errors"
	"strings"

	"github.com/Shadowmaple/logflow/field"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/model"
	"github.com/Shadowmaple/logflow/render"
)

type UppercaseFilter struct {
	fields map[field.FieldSetter]render.Render
}

func init() {
	register("uppercase", newUppercaseFilter)
}

func newUppercaseFilter(config map[string]any) model.Filter {
	fields, ok := config["fields"].([]string)
	if !ok {
		panic("uppercase invalid fields")
	}
	mp := make(map[field.FieldSetter]render.Render)
	for _, f := range fields {
		mp[field.NewFieldSetter(f, true)] = render.GetRender(f)
	}
	return &UppercaseFilter{fields: mp}
}

// 若遇到一个字段处理失败，则继续处理其它的，最后返回错误
func (f *UppercaseFilter) Filter(event *event.Event) error {
	var failed bool
	for fieldSetter, render := range f.fields {
		if value, err := render.Render(event); err == nil {
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
		return errors.New("uppercase filter failed")
	}
	return nil
}
