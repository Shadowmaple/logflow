package filter

import (
	"encoding/json"
	"errors"
	"maps"

	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"
	"github.com/Shadowmaple/logflow/internal/utils"

	"go.uber.org/zap"
)

type JsonFilter struct {
	Source string
	Target string
}

func (jf *JsonFilter) Filter(event *model.Event) (err error) {
	// 从source解析json，并将结果存储到target
	raw, ok := event.Data[jf.Source]
	if !ok {
		logger.Error("json filter: source not found")
		return errors.New("source not found")
	}
	// 判断类型是否为string
	if ok := utils.TypeCheck(raw, "string"); !ok {
		logger.Error("json filter: source is not a string")
		return errors.New("invalid source")
	}
	rawStr := raw.(string)
	// 解析json
	var data map[string]any
	if err = json.Unmarshal([]byte(rawStr), &data); err != nil {
		logger.Error("json filter: decode failed:"+err.Error(), zap.String("data", rawStr))
		return errors.New("JSON decode failed")
	}
	// 将解析结果存储到target
	if jf.Target != "" {
		event.Data[jf.Target] = data
	} else {
		// 若target为空，则将解析结果更新至event.Data
		maps.Copy(event.Data, data)
	}
	return nil
}

func NewJsonFilter(conf map[string]any) Filter {
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
	register("json", NewJsonFilter)
}
