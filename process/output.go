package process

import (
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"
	"github.com/Shadowmaple/logflow/output"
	"go.uber.org/zap"
)

type OutputProcessor struct {
	output.Output
}

func BuildOutputProcessors(confs []map[string]any) []*OutputProcessor {
	outputs := output.BuildOutputs(confs)
	l := make([]*OutputProcessor, 0, len(outputs))
	for _, o := range outputs {
		l = append(l, &OutputProcessor{Output: o})
	}
	return l
}

func (o *OutputProcessor) Process(event *model.Event) bool {
	if err := o.Output.Handle(event); err != nil {
		logger.Error("output process event failed", zap.Error(err))
		return false
	}
	return true
}

func (o *OutputProcessor) Close() {
	o.Output.Close()
}
