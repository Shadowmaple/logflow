package render

import (
	"fmt"

	"github.com/Shadowmaple/logflow/internal/event"
)

// MultiLevelRender 多层级解析器，如[@metadata][kafka][topic]
type MultiLevelRender struct {
	preFields []string
	lastField string
}

// func newMultiLevelRender(s string) *MultiLevelRender {
// 	return &MultiLevelRender{fields: getMultiLevelFields(s)}
// }

func newMultiLevelRender(fields []string) *MultiLevelRender {
	return &MultiLevelRender{preFields: fields[:len(fields)-1], lastField: fields[len(fields)-1]}
}

// getMultiLevelFields 解析多层级字段，如[@metadata][kafka][topic]，返回["@metadata", "kafka", "topic"]
// func getMultiLevelFields(s string) []string {
// 	return strings.Split(strings.TrimSuffix(strings.TrimPrefix(s, "%{["), "]}"), "][")
// }

// Render 解析事件数据，返回解析后的值
func (r *MultiLevelRender) Render(event *event.Event) (any, error) {
	var cur map[string]any = event.Data
	// 逐层获取fields的值
	for _, field := range r.preFields {
		if cur == nil {
			return nil, fmt.Errorf("multi level render failed, %s not found", field)
		}
		v, ok := cur[field]
		if !ok {
			return nil, fmt.Errorf("multi level render failed, %s not found", field)
		}
		cur, ok = v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("multi level render failed, %s not found", field)
		}
	}
	if cur == nil {
		return nil, fmt.Errorf("multi level render failed, %s not found", r.lastField)
	}
	if val, ok := cur[r.lastField]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("multi level render failed, %s not found", r.lastField)
}
