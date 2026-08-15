package field_getter

import (
	"github.com/Shadowmaple/logflow/internal/event"
)

// MultiLevelFieldGetter 多层级解析器，如[@metadata][kafka][topic]
type MultiLevelFieldGetter struct {
	preFields []string
	lastField string
}

// func newMultiLevelFieldGetter(s string) *MultiLevelFieldGetter {
// 	return &MultiLevelFieldGetter{fields: getMultiLevelFields(s)}
// }

func newMultiLevelFieldGetter(fields []string) *MultiLevelFieldGetter {
	return &MultiLevelFieldGetter{preFields: fields[:len(fields)-1], lastField: fields[len(fields)-1]}
}

// getMultiLevelFields 解析多层级字段，如[@metadata][kafka][topic]，返回["@metadata", "kafka", "topic"]
// func getMultiLevelFields(s string) []string {
// 	return strings.Split(strings.TrimSuffix(strings.TrimPrefix(s, "%{["), "]}"), "][")
// }

// GetField 解析事件数据，返回解析后的值
func (r *MultiLevelFieldGetter) GetField(event *event.Event) (any, error) {
	var cur map[string]any = event.Data
	// 逐层获取fields的值
	for _, field := range r.preFields {
		if cur == nil {
			return nil, ErrNotFound
		}
		v, ok := cur[field]
		if !ok {
			return nil, ErrNotFound
		}
		cur, ok = v.(map[string]any)
		if !ok {
			return nil, ErrInvalidType
		}
	}
	if cur == nil {
		return nil, ErrNotFound
	}
	if val, ok := cur[r.lastField]; ok {
		return val, nil
	}
	return nil, ErrNotFound
}
