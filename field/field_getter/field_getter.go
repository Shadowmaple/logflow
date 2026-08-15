package field_getter

import (
	"errors"
	"regexp"

	"github.com/Shadowmaple/logflow/internal/event"
)

var ErrNotFound = errors.New("field not found")
var ErrInvalidType = errors.New("field is not a valid type")

var matchp, matchESIndex, jsonPath *regexp.Regexp

func init() {
	matchp, _ = regexp.Compile(`^(\[.*?\])+$`) // []解析模板，如["a"], ["a"]["b"]
	// matchGoTemp, _ = regexp.Compile(`{{.*}}`)
	matchESIndex, _ = regexp.Compile(`%{.*?}`) //%{+YYYY.MM.dd}
	jsonPath, _ = regexp.Compile(`^\$\.`)
}

type FieldGetter interface {
	GetField(*event.Event) (any, error)
}

// 获取字段取值获取器，如未匹配到则默认为单层级获取器
func GetFieldGetter(template string) FieldGetter {
	fieldgetter := getValueFieldGetter(template)
	if fieldgetter != nil {
		return fieldgetter
	}
	return newOneLevelFieldGetter(template)
}

// 获取字段取值获取器，如未匹配到则默认为字面量解析器
func GetFieldGetter2(template string) FieldGetter {
	fieldgetter := getValueFieldGetter(template)
	if fieldgetter != nil {
		return fieldgetter
	}
	return newNormalFieldGetter(template)
}

// getValueFieldGetter matches all regexp pattern and return a ValueFieldGetter
// return nil if no pattern matched
func getValueFieldGetter(template string) FieldGetter {
	if matchp.Match([]byte(template)) {
		findp, _ := regexp.Compile(`(\[(.*?)\])`)
		fields := make([]string, 0)
		for _, v := range findp.FindAllStringSubmatch(template, -1) {
			fields = append(fields, v[2])
		}

		if len(fields) == 1 {
			return newOneLevelFieldGetter(fields[0])
		}
		return newMultiLevelFieldGetter(fields)
	}
	if matchESIndex.Match([]byte(template)) {
		return newIndexFieldGetter(template)
	}
	if jsonPath.Match([]byte(template)) {
		return newJsonpathFieldGetter(template)
	}

	return nil
}
