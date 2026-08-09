package render

import (
	"regexp"

	"github.com/Shadowmaple/logflow/internal/event"
)

var matchp, matchESIndex, jsonPath *regexp.Regexp

func init() {
	matchp, _ = regexp.Compile(`^(\[.*?\])+$`) // []解析模板，如["a"], ["a"]["b"]
	// matchGoTemp, _ = regexp.Compile(`{{.*}}`)
	matchESIndex, _ = regexp.Compile(`%{.*?}`) //%{+YYYY.MM.dd}
	jsonPath, _ = regexp.Compile(`^\$\.`)
}

type Render interface {
	Render(*event.Event) (any, error)
}

// 获取字段取值渲染器，如未匹配到则默认为单层级渲染器
func GetRender(template string) Render {
	render := getValueRender(template)
	if render != nil {
		return render
	}
	return newOneLevelRender(template)
}

// getValueRender matches all regexp pattern and return a ValueRender
// return nil if no pattern matched
func getValueRender(template string) Render {
	if matchp.Match([]byte(template)) {
		findp, _ := regexp.Compile(`(\[(.*?)\])`)
		fields := make([]string, 0)
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}

		if len(fields) == 1 {
			return newOneLevelRender(fields[0])
		}
		return newMultiLevelRender(fields)
	}
	if matchESIndex.Match([]byte(template)) {
		return newIndexRender(template)
	}
	if jsonPath.Match([]byte(template)) {
		return newJsonpathRender(template)
	}

	return nil
}
