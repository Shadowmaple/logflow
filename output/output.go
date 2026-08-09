package output

import (
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
)

func BuildOutputs(configs []map[string]any) []model.Output {
	outputs := make([]model.Output, 0, len(configs))
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
