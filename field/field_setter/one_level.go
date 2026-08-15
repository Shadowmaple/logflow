package field_setter

import "github.com/Shadowmaple/logflow/internal/event"

type OneLevelFieldSetter struct {
	field     string
	overwrite bool
}

func newOneLevelFieldSetter(field string, overwrite bool) *OneLevelFieldSetter {
	return &OneLevelFieldSetter{
		field:     field,
		overwrite: overwrite,
	}
}

func (f *OneLevelFieldSetter) SetField(event *event.Event, value any) {
	if event == nil {
		return
	}
	if _, ok := event.Data[f.field]; !ok || f.overwrite {
		event.Data[f.field] = value
		return
	}
}
