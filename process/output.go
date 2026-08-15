package process

import (
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
	"github.com/Shadowmaple/logflow/output"

	"go.uber.org/zap"
)

type OutputProcessor struct {
	model.Output
}

func BuildOutputProcessors(confs []map[any]any) []*OutputProcessor {
	outputs := output.BuildOutputs(confs)
	l := make([]*OutputProcessor, 0, len(outputs))
	for _, o := range outputs {
		l = append(l, &OutputProcessor{Output: o})
	}
	return l
}

func (o *OutputProcessor) Process(event *event.Event) *event.Event {
	if err := o.Output.Handle(event); err != nil {
		logger.Error("output process event failed", zap.Error(err))
		return event
	}
	return event
}

func (o *OutputProcessor) Close() {
	o.Output.Close()
}
