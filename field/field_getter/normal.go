package field_getter

import "github.com/Shadowmaple/logflow/internal/event"

// 字面量值渲染器，即返回值本身
type NormalFieldGetter struct {
	value string
}

// NewNormalFieldGetter creates a new NormalFieldGetter instance
func newNormalFieldGetter(template string) *NormalFieldGetter {
	return &NormalFieldGetter{value: template}
}

// GetField returns the value
func (r *NormalFieldGetter) GetField(event *event.Event) (any, error) {
	return r.value, nil
}
