package model

import (
	"github.com/Shadowmaple/logflow/condition"
	"github.com/Shadowmaple/logflow/field/field_setter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"

	"go.uber.org/zap"
)

type Filter interface {
	Filter(event *event.Event) (*event.Event, error)
}

type FilterProcessor struct {
	Filter

	condition *condition.ConditionFilter
	config    map[string]any
	failTag   string
	addFields map[field_setter.FieldSetter]any
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
		if addFieldAny, ok := vConf["add_fields"]; ok {
			addFields, ok := addFieldAny.(map[string]any)
			if !ok {
				logger.Error("add_fields config is not map[string]any", zap.Any("config", addFieldAny))
				continue
			}
			p.addFields = make(map[field_setter.FieldSetter]any, len(addFields))
			for k, v := range addFields {
				p.addFields[field_setter.NewFieldSetter(k, false)] = v
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

func (f *FilterProcessor) Process(event *event.Event) *event.Event {
	var err error
	if event != nil && f.condition.Check(event) {
		if event, err = f.Filter.Filter(event); err != nil {
			logger.Error("filter process event failed", zap.Error(err))
			if f.failTag != "" {
				event.Data["@fail_tag"] = f.failTag
			}
			return event
		}
		if f.addFields != nil {
			for k, v := range f.addFields {
				k.SetField(event, v)
			}
		}
	}
	return event
}

type FilterProcessGroup []*FilterProcessor

func (f *FilterProcessGroup) Process(event *event.Event) *event.Event {
	for _, p := range *f {
		event = p.Process(event)
	}
	return event
}
