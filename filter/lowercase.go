package filter

import (
	"errors"
	"strings"

	"github.com/Shadowmaple/logflow/field/field_getter"
	"github.com/Shadowmaple/logflow/field/field_setter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/model"
)

type LowercaseFilter struct {
	fields map[field_setter.FieldSetter]field_getter.FieldGetter
}

func init() {
	register("lowercase", newLowercaseFilter)
}

func newLowercaseFilter(config map[any]any) model.Filter {
	fields, ok := config["fields"].([]any)
	if !ok {
		panic("lowercase invalid fields")
	}
	mp := make(map[field_setter.FieldSetter]field_getter.FieldGetter)
	for _, f := range fields {
		nameStr := f.(string)
		mp[field_setter.NewFieldSetter(nameStr, true)] = field_getter.GetFieldGetter(nameStr)
	}
	return &LowercaseFilter{fields: mp}
}

// 若遇到一个字段处理失败，则继续处理其它的，最后返回错误
func (f *LowercaseFilter) Filter(event *event.Event) (*event.Event, error) {
	var failed bool
	for fieldSetter, fieldgetter := range f.fields {
		if value, err := fieldgetter.GetField(event); err == nil {
			v, ok := value.(string)
			if !ok {
				failed = true
				continue
			}
			fieldSetter.SetField(event, strings.ToLower(v))
		} else if !errors.Is(err, field_getter.ErrNotFound) {
			failed = true
		}
	}
	if failed {
		return event, errors.New("lowercase filter failed")
	}
	return event, nil
}
