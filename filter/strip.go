package filter

import (
	"errors"
	"strings"

	"github.com/Shadowmaple/logflow/field/field_getter"
	"github.com/Shadowmaple/logflow/field/field_setter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
)

// 去除值的首尾空格、换行符、制表符等无效字符
type StripFilter struct {
	fields map[field_setter.FieldSetter]field_getter.FieldGetter
}

func init() {
	register("strip", newStripFilter)
}

func newStripFilter(conf map[any]any) model.Filter {
	if conf == nil {
		logger.Warn("strip filter config is nil")
		return nil
	}
	fieldAny, ok := conf["fields"]
	if !ok {
		logger.Warn("strip filter config is missing fields expression")
		return nil
	}
	fields, ok := fieldAny.([]any)
	if !ok {
		logger.Error("strip filter config fields expression type is not array")
		return nil
	}
	res := make(map[field_setter.FieldSetter]field_getter.FieldGetter, len(fields))
	for _, name := range fields {
		nameStr := name.(string)
		res[field_setter.NewFieldSetter(nameStr, true)] = field_getter.GetFieldGetter(nameStr)
	}
	return &StripFilter{
		fields: res,
	}
}

func (f *StripFilter) Filter(event *event.Event) (*event.Event, error) {
	var failed bool
	for fieldSetter, getter := range f.fields {
		if value, err := getter.GetField(event); err == nil {
			v, ok := value.(string)
			if !ok {
				failed = true
				continue
			}
			fieldSetter.SetField(event, strings.TrimSpace(v))
		} else if !errors.Is(err, field_getter.ErrNotFound) {
			failed = true
		}
	}
	if failed {
		return nil, errors.New("strip filter failed")
	}
	return event, nil
}
