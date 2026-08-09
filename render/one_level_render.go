package render

import (
	"fmt"

	"github.com/Shadowmaple/logflow/internal/event"
)

type OneLevelRender struct {
	field string
}

func newOneLevelRender(s string) *OneLevelRender {
	return &OneLevelRender{field: s}
}

// Render 解析事件数据，返回解析后的值
func (r *OneLevelRender) Render(event *event.Event) (any, error) {
	if val, ok := event.Data[r.field]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("one level render failed, %s not found", r.field)
}
