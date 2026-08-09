package render

import "github.com/Shadowmaple/logflow/internal/event"

// 字面量值渲染器，即返回值本身
type NormalRender struct {
	value any
}

// NewNormalRender creates a new NormalRender instance
func newNormalRender(template any) *NormalRender {
	return &NormalRender{value: template}
}

// Render renders the value
func (r *NormalRender) Render(e *event.Event) (any, error) {
	return r.value, nil
}
