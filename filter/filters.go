package filter

import (
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
)

// 置于 if 条件判断下的多个 filter
type FiltersFilter struct {
	processor *model.ProcessNode
}

func init() {
	register("filters", newFiltersFilter)
}

func newFiltersFilter(conf map[string]any) model.Filter {
	if conf == nil {
		logger.Warn("filters filter config is nil")
		return nil
	}
	// 获取其它modules的配置
	filtersConf, ok := conf["modules"].([]map[string]any)
	if !ok {
		logger.Warn("filters modules config is not a list")
		return nil
	}
	// 创建 filters 处理节点
	processors := model.BuildFilterProcessors(filtersConf, BuildFilter)

	// 构建处理链表
	var processorHead *model.ProcessNode
	for _, filter := range processors {
		processorHead = model.AppendProcessors(processorHead, filter)
	}
	return &FiltersFilter{
		processor: processorHead,
	}
}

func (f *FiltersFilter) Filter(event *event.Event) (*event.Event, error) {
	event = f.processor.Process(event)
	return event, nil
}
