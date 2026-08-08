package process

import (
	commonFilter "github.com/Shadowmaple/logflow/common_filter"
	"github.com/Shadowmaple/logflow/filter"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"

	"go.uber.org/zap"
)

type FilterProcessor struct {
	filter.Filter

	condition *commonFilter.ConditionFilter
	config    map[string]any
	failTag   string
	// removeFields []field_deleter.FieldDeleter
	// addFields    map[field_setter.FieldSetter]value_render.ValueRender
}

func buildFilterProcessor(conf map[string]any) *FilterProcessor {
	filter := filter.BuildFilter(conf)
	if filter == nil {
		return nil
	}
	p := &FilterProcessor{
		Filter: filter,
		config: conf,
	}
	for _, v := range conf {
		vConf := v.(map[string]any)
		p.condition = commonFilter.NewConditionFilter(vConf)
		if failTag, ok := vConf["fail_tag"]; ok {
			p.failTag = failTag.(string)
		}
	}
	return p
}

func BuildFilterProcessors(confs []map[string]any) []*FilterProcessor {
	filters := filter.BuildFilters(confs)
	res := make([]*FilterProcessor, len(filters))
	for i, f := range filters {
		res[i] = buildFilterProcessor(confs[i])
		res[i].Filter = f
	}
	return res
}

func (f *FilterProcessor) Process(event *model.Event) bool {
	if f.condition.Check(event) {
		if err := f.Filter.Filter(event); err != nil {
			logger.Error("filter process event failed", zap.Error(err))
			return false
		}
	}
	return true
}
