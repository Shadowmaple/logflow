package filter

import (
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
)

func BuildFilters(conf []map[any]any) []model.Filter {
	if conf == nil {
		return nil
	}
	filters := make([]model.Filter, 0, len(conf))
	for _, c := range conf {
		f := BuildFilter(c)
		if f != nil {
			filters = append(filters, f)
		}
	}
	return filters
}

func BuildFilter(conf map[any]any) model.Filter {
	if conf == nil {
		return nil
	}
	for k, v := range conf {
		filterType, filterConf := k.(string), v.(map[any]any)
		if handler, ok := filterHandlers[filterType]; ok {
			return handler(filterConf)
		}
		logger.Error("filter config type not found: " + filterType)
	}
	return nil
}

func BuildFilterByType(filterType string, conf map[any]any) model.Filter {
	if conf == nil {
		return nil
	}
	if handler, ok := filterHandlers[filterType]; ok {
		return handler(conf)
	}
	logger.Error("filter config type not found: " + filterType)
	return nil
}

var filterHandlers = make(map[string]func(conf map[any]any) model.Filter)

func register(name string, f func(conf map[any]any) model.Filter) {
	filterHandlers[name] = f
}
