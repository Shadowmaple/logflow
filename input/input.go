package input

import "github.com/Shadowmaple/logflow/model"

func NewInput(conf map[string]any) model.Input {
	if conf == nil {
		panic("input config is nil")
	}
	for k, v := range conf {
		switch k {
		case "kafka":
			return newKafkaInput(v.(map[string]any))
		}
		panic("invalid input type: " + k)
	}
	return nil
}
