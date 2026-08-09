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
	// 获取其它filters的配置
	filtersConf, ok := conf["filters"].([]map[string]any)
	if !ok {
		logger.Warn("filters filter config is not a list")
		return nil
	}
	// 创建 filters 处理节点
	processors := model.BuildFilterProcessors(filtersConf, BuildFilter)

	// 构建处理链表
	processorNode := model.NewProcessList()
	for _, filter := range processors {
		processorNode.Append(filter)
	}
	return &FiltersFilter{
		processor: processorNode,
	}
}

func (f *FiltersFilter) Filter(event *event.Event) error {
	f.processor.Process(event)
	return nil
}
