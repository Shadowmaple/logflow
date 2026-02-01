package filter

import "github.com/Shadowmaple/logflow/internal/model"

type Filter interface {
	Filter(event *model.Event) error
}

func BuildFilters(conf []map[string]any) []Filter {
	if conf == nil {
		return nil
	}
	filters := make([]Filter, 0, len(conf))
	for _, c := range conf {
		for k, v := range c {
			if handler, ok := filterHandlers[k]; ok {
				filters = append(filters, handler(v.(map[string]any)))
			}
		}
	}
	return filters
}

var filterHandlers = make(map[string]func(conf map[string]any) Filter)

func register(name string, f func(conf map[string]any) Filter) {
	filterHandlers[name] = f
}
