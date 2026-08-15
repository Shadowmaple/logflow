package output

import (
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
)

var outputHandlers = make(map[string]func(map[any]any) model.Output)

func Register(name string, handler func(map[any]any) model.Output) {
	outputHandlers[name] = handler
}

func BuildOutputs(configs []map[string]any) []model.Output {
	outputs := make([]model.Output, 0, len(configs))
	for _, config := range configs {
		for k, v := range config {
			if handler, ok := outputHandlers[k]; ok {
				outputs = append(outputs, handler(v.(map[any]any)))
			} else {
				logger.Fatal("output: unknown type " + k)
			}
		}
	}
	return outputs
}
