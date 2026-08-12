package field_setter
import (
	"fmt"

	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
)

type MultiLevelFieldSetter struct {
	preFields []string
	lastField string
	overwrite bool
}

func newMultiLevelFieldSetter(fields []string, overwrite bool) *MultiLevelFieldSetter {
	return &MultiLevelFieldSetter{
		preFields: fields[:len(fields)-1],
		lastField: fields[len(fields)-1],
		overwrite: overwrite,
	}
}

// 如果不存在该层级的field，则创建一个新的map写入
// 若非最后一层的field不是map，则跳过。TODO：是否需要判断overwrite
func (f *MultiLevelFieldSetter) SetField(event *event.Event, value any) {
	var cur = event.Data
	for _, pre := range f.preFields {
		if _, ok := cur[pre]; !ok {
			t := make(map[string]any)
			cur[pre] = t
			cur = t
		} else {
			cur, ok = cur[pre].(map[string]any)
			if !ok {
				logger.Error(fmt.Sprintf("field %s is not a map", pre))
				return
			}
		}
	}
	if _, ok := cur[f.lastField]; !ok || f.overwrite {
		cur[f.lastField] = value
		return
	}
}
