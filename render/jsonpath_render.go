package render

import (
	"fmt"

	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/oliveagle/jsonpath"
)

// JSON 路径解析器，如 $.data.name
type JsonpathRender struct {
	Pat *jsonpath.Compiled
}

func newJsonpathRender(path string) *JsonpathRender {
	pat, err := jsonpath.Compile(path)
	if err != nil {
		panic(fmt.Sprintf("json path compile `%s` error: %s", path, err))
	}
	return &JsonpathRender{Pat: pat}
}

func (r *JsonpathRender) Render(event *event.Event) (any, error) {
	return r.Pat.Lookup(event.Data)
}
