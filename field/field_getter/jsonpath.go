package field_getter

import (
	"fmt"

	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/oliveagle/jsonpath"
)

// JSON 路径解析器，如 $.data.name
type JsonpathFieldGetter struct {
	Pat *jsonpath.Compiled
}

func newJsonpathFieldGetter(path string) *JsonpathFieldGetter {
	pat, err := jsonpath.Compile(path)
	if err != nil {
		panic(fmt.Sprintf("json path compile `%s` error: %s", path, err))
	}
	return &JsonpathFieldGetter{Pat: pat}
}

func (r *JsonpathFieldGetter) GetField(event *event.Event) (any, error) {
	return r.Pat.Lookup(event.Data)
}
