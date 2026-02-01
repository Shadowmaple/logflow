package render

import "github.com/Shadowmaple/logflow/internal/model"

type Render interface {
	Render(*model.Event) (any, error)
}

func GetRender(name string, template string) Render {
	switch name {
	case "index":
		return newIndexRender(template)
	default:
		return nil
	}
}
