package output

import (
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"
)

type Output interface {
	Handle(event *model.Event) error
	Close()
}

func BuildOutputs(configs []map[string]any) []Output {
	outputs := make([]Output, 0, len(configs))
	for _, config := range configs {
		for k, v := range config {
			switch k {
			case "elasticsearch":
				outputs = append(outputs, newElasticsearchOutput(v.(map[string]any)))
			default:
				logger.Fatal("output: unknown type " + k)
			}
		}
	}
	return outputs
}
