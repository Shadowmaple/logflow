package model

import (
	"github.com/Shadowmaple/logflow/condition"
	"github.com/Shadowmaple/logflow/field"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"

	"go.uber.org/zap"
)

type Filter interface {
	Filter(event *event.Event) error
}

type FilterProcessor struct {
	Filter

	condition *condition.ConditionFilter
	config    map[string]any
	failTag   string
	addFields map[field.FieldSetter]any
	// removeFields []field_deleter.FieldDeleter
}

type buildFilterFunc func(config map[string]any) Filter

func buildFilterProcessor(conf map[string]any, buildFilterFunc buildFilterFunc) *FilterProcessor {
	filter := buildFilterFunc(conf)
	if filter == nil {
		logger.Fatal("filter processor build failed", zap.Any("config", conf))
		return nil
	}
	p := &FilterProcessor{
		Filter: filter,
		config: conf,
	}
	for _, v := range conf {
		vConf := v.(map[string]any)
		p.condition = condition.NewConditionFilter(vConf)
		if failTag, ok := vConf["fail_tag"]; ok {
			p.failTag = failTag.(string)
		}
		if addFields, ok := vConf["add_fields"]; ok {
			// p.addFields = make(map[field.FieldSetter]render.Render)
			for k, v := range addFields.(map[string]any) {
				p.addFields[field.NewFieldSetter(k, false)] = v
			}
		}
	}
	return p
}

func BuildFilterProcessors(confs []map[string]any, buildFilterFunc buildFilterFunc) []*FilterProcessor {
	res := make([]*FilterProcessor, len(confs))
	for i, conf := range confs {
		res[i] = buildFilterProcessor(conf, buildFilterFunc)
	}
	return res
}

func (f *FilterProcessor) Process(event *event.Event) bool {
	if event != nil && f.condition.Check(event) {
		if err := f.Filter.Filter(event); err != nil {
			logger.Error("filter process event failed", zap.Error(err))
			if f.failTag != "" {
				event.Data["fail_tag"] = f.failTag
			}
			return false
		}
		if f.addFields != nil {
			for k, v := range f.addFields {
				k.SetField(event, v)
			}
		}
	}
	return true
}
