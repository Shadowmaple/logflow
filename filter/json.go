package filter

import (
	"encoding/json"
	"errors"
	"maps"

	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/utils"
	"github.com/Shadowmaple/logflow/model"

	"go.uber.org/zap"
)

// TODO: target 支持多层级字段
type JsonFilter struct {
	Source string
	Target string
}

func (jf *JsonFilter) Filter(event *event.Event) (*event.Event, error) {
	// 从source解析json，并将结果存储到target
	raw, ok := event.Data[jf.Source]
	if !ok {
		logger.Warn("json filter: source not found")
		return event, errors.New("source not found")
	}
	// 判断类型是否为[]byte
	rawStr, ok := utils.ParseToBytes(raw)
	if !ok {
		logger.Warn("json filter: source is not a []byte", zap.Any("value", raw))
		return event, errors.New("invalid source")
	}
	// 解析json
	var data map[string]any
	if err := json.Unmarshal(rawStr, &data); err != nil {
		logger.Warn("json filter: decode failed:"+err.Error(), zap.String("data", string(rawStr)))
		return event, errors.New("JSON decode failed")
	}
	// 将解析结果存储到target
	if jf.Target != "" {
		event.Data[jf.Target] = data
	} else {
		// 若target为空，则将解析结果更新至event.Data
		maps.Copy(event.Data, data)
	}
	return event, nil
}

func newJsonFilter(conf map[any]any) model.Filter {
	f := &JsonFilter{
		Target: "",
	}
	if v, ok := conf["source"]; ok {
		f.Source = v.(string)
	} else {
		panic("json filter: source is required")
	}
	if v, ok := conf["target"]; ok {
		f.Target = v.(string)
	}
	return f
}

func init() {
	register("json", newJsonFilter)
}
