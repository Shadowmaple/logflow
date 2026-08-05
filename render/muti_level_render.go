package render

import (
	"fmt"
	"strings"

	"github.com/Shadowmaple/logflow/internal/model"
)

// MultiLevelRender 多层级解析器，如[@metadata][kafka][topic]
type MultiLevelRender struct {
	fields []string
}

func newMultiLevelRender(s string) *MultiLevelRender {
	return &MultiLevelRender{fields: getMultiLevelFields(s)}
}

// getMultiLevelFields 解析多层级字段，如[@metadata][kafka][topic]，返回["@metadata", "kafka", "topic"]
func getMultiLevelFields(s string) []string {
	return strings.Split(strings.TrimSuffix(strings.TrimPrefix(s, "%{["), "]}"), "][")
}

// Render 解析事件数据，返回解析后的值
func (r *MultiLevelRender) Render(event *model.Event) (any, error) {
	var (
		cur any = event.Data
		tmp map[string]any
		ok  bool
	)
	// 逐层获取fields的值
	for _, field := range r.fields {
		if cur == nil {
			return nil, fmt.Errorf("multi level render failed, %s not found", field)
		}
		tmp, ok = cur.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("multi level render failed, %s not found", field)
		}
		cur, ok = tmp[field]
		if !ok {
			return nil, fmt.Errorf("multi level render failed, %s not found", field)
		}
	}
	return cur, nil
}
