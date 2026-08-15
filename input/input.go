package input

import (
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
)

type InputHander func(conf map[any]any) model.Input

var inputHandlers = make(map[string]InputHander)

func Register(name string, f InputHander) {
	inputHandlers[name] = f
}

func NewInput(conf map[any]any) model.Input {
	if conf == nil {
		panic("input config is nil")
	}
	for k, v := range conf {
		if handler, ok := inputHandlers[k.(string)]; ok {
			return handler(v.(map[any]any))
		}
		logger.Fatal("invalid input type: " + k.(string))
	}
	return nil
}
