package process

import (
	"github.com/Shadowmaple/logflow/filter"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"

	"go.uber.org/zap"
)

type FilterProcessor struct {
	filter.Filter
}

func BuildFilterProcessors(confs []map[string]any) []*FilterProcessor {
	filters := filter.BuildFilters(confs)
	l := make([]*FilterProcessor, 0, len(filters))
	for _, f := range filters {
		l = append(l, &FilterProcessor{Filter: f})
	}
	return l
}

func (f *FilterProcessor) Process(event *model.Event) bool {
	if err := f.Filter.Filter(event); err != nil {
		logger.Error("filter process event failed", zap.Error(err))
		return false
	}
	return true
}
