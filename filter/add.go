package filter

import (
	"github.com/Shadowmaple/logflow/field/field_setter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/utils"
	"github.com/Shadowmaple/logflow/model"
)

type addField struct {
	setter field_setter.FieldSetter
	value  any
}

// 添加字段
type AddFilter struct {
	fields []addField
}

func init() {
	register("add", newAddFilter)
}

func newAddFilter(conf map[any]any) model.Filter {
	if conf == nil {
		logger.Warn("add filter config is nil")
		return nil
	}

	overwrite := true
	if v, ok := conf["overwrite"]; ok {
		overwrite, ok = utils.ParseToBool(v)
		if !ok {
			logger.Error("add filter config overwrite type is invalid")
			return nil
		}
	}

	fieldsAny, ok := conf["fields"]
	if !ok {
		logger.Warn("add filter config is missing fields")
		return nil
	}
	fields, ok := fieldsAny.(map[any]any)
	if !ok {
		logger.Error("add filter config fields type is not a map")
		return nil
	}
	if len(fields) == 0 {
		logger.Warn("add filter config fields is empty")
		return nil
	}

	res := make([]addField, 0, len(fields))
	for name, value := range fields {
		nameStr, ok := name.(string)
		if !ok {
			logger.Error("add filter config fields key type is not string")
			return nil
		}
		res = append(res, addField{
			setter: field_setter.NewFieldSetter(nameStr, overwrite),
			value:  value,
		})
	}
	return &AddFilter{fields: res}
}

func (f *AddFilter) Filter(event *event.Event) (*event.Event, error) {
	for _, field := range f.fields {
		field.setter.SetField(event, field.value)
	}
	return event, nil
}
