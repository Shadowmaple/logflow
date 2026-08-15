package field_getter

import (
	"fmt"

	"github.com/Shadowmaple/logflow/internal/event"
)

type OneLevelFieldGetter struct {
	field string
}

func newOneLevelFieldGetter(s string) *OneLevelFieldGetter {
	return &OneLevelFieldGetter{field: s}
}

// GetField 解析事件数据，返回解析后的值
func (r *OneLevelFieldGetter) GetField(event *event.Event) (any, error) {
	if val, ok := event.Data[r.field]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("one level fieldgetter failed, %s not found", r.field)
}
