package filter

import (
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"

	"go.uber.org/zap"
)

type Filter interface {
	Filter(event *model.Event) error
}

func BuildFilters(conf []map[string]any) []Filter {
	if conf == nil {
		return nil
	}
	filters := make([]Filter, 0, len(conf))
	for _, c := range conf {
		f := BuildFilter(c)
		if f != nil {
			filters = append(filters, f)
		}
	}
	return filters
}

func BuildFilter(conf map[string]any) Filter {
	if conf == nil {
		return nil
	}
	for k, v := range conf {
		filterType, filterConf := k, v.(map[string]any)
		logger.Info("filter config type: "+filterType, zap.Any("conf", filterConf))
		if handler, ok := filterHandlers[filterType]; ok {
			return handler(filterConf)
		}
		logger.Error("filter config type not found: " + filterType)
	}
	return nil
}

var filterHandlers = make(map[string]func(conf map[string]any) Filter)

func register(name string, f func(conf map[string]any) Filter) {
	filterHandlers[name] = f
}
