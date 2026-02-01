package input

import (
	"github.com/Shadowmaple/logflow/internal/model"
)

type Input interface {
	// Receive() <-chan *model.Event
	ReceiveOne() *model.Event
	Close()
}

func NewInput(conf map[string]any) Input {
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
