package filter

import (
	"strings"

	"github.com/Shadowmaple/logflow/field/field_setter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
	"github.com/Shadowmaple/logflow/render"
)

// 去除值的首尾空格、换行符、制表符等无效字符
type StripFilter struct {
	fields map[field_setter.FieldSetter]render.Render
}

func init() {
	register("strip", newStripFilter)
}

func newStripFilter(conf map[string]any) model.Filter {
	if conf == nil {
		logger.Warn("strip filter config is nil")
		return nil
	}
	fieldAny, ok := conf["fields"]
	if !ok {
		logger.Warn("strip filter config is missing fields expression")
		return nil
	}
	fields, ok := fieldAny.([]string)
	if !ok {
		logger.Error("strip filter config fields expression type is not array")
		return nil
	}
	res := make(map[field_setter.FieldSetter]render.Render, len(fields))
	for _, name := range fields {
		res[field_setter.NewFieldSetter(name, true)] = render.GetRender(name)
	}
	return &StripFilter{
		fields: res,
	}
}

func (f *StripFilter) Filter(event *event.Event) (*event.Event, error) {
	var failed bool
	for fieldSetter, render := range f.fields {
		if value, err := render.Render(event); err == nil {
			v, ok := value.(string)
			if !ok {
				failed = true
				continue
			}
			fieldSetter.SetField(event, strings.TrimSpace(v))
		} else {
			failed = true
		}
	}
	if failed {
		logger.Error("strip filter failed")
	}
	return event, nil
}
