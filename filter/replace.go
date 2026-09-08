package filter

import (
	"errors"
	"strings"

	"github.com/Shadowmaple/logflow/field/field_getter"
	"github.com/Shadowmaple/logflow/field/field_setter"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/utils"
	"github.com/Shadowmaple/logflow/model"
)

type replaceRule struct {
	getter field_getter.FieldGetter
	setter field_setter.FieldSetter
	old    string
	new    string
	count  int // 替换次数, 小于 0 表示替换所有
}

// 字符串替换
type ReplaceFilter struct {
	fields []replaceRule
}

func init() {
	register("replace", newReplaceFilter)
}

func newReplaceFilter(conf map[any]any) model.Filter {
	if conf == nil {
		logger.Warn("replace filter config is nil")
		return nil
	}

	fieldsAny, ok := conf["fields"]
	if !ok {
		logger.Warn("replace filter config is missing fields")
		return nil
	}
	fields, ok := fieldsAny.(map[any]any)
	if !ok {
		logger.Error("replace filter config fields type is not a map")
		return nil
	}
	if len(fields) == 0 {
		logger.Warn("replace filter config fields is empty")
		return nil
	}

	res := make([]replaceRule, 0, len(fields))
	for name, ruleAny := range fields {
		nameStr, ok := name.(string)
		if !ok {
			logger.Error("replace filter config fields key type is not string")
			return nil
		}
		rule, ok := parseReplaceRule(nameStr, ruleAny)
		if !ok {
			return nil
		}
		res = append(res, rule)
	}
	return &ReplaceFilter{fields: res}
}

// 规则格式: [old, new] 或 [old, new, count]，count 不给则替换所有
func parseReplaceRule(name string, ruleAny any) (replaceRule, bool) {
	rules, ok := ruleAny.([]any)
	if !ok {
		logger.Error("replace filter config rule type is not array")
		return replaceRule{}, false
	}
	if len(rules) < 2 || len(rules) > 3 {
		logger.Error("replace filter config rule length must be 2 or 3")
		return replaceRule{}, false
	}
	old, ok := rules[0].(string)
	if !ok {
		logger.Error("replace filter config rule old value type is not string")
		return replaceRule{}, false
	}
	newStr, ok := rules[1].(string)
	if !ok {
		logger.Error("replace filter config rule new value type is not string")
		return replaceRule{}, false
	}
	count := -1
	if len(rules) == 3 {
		count, ok = utils.ParseToInt(rules[2])
		if !ok {
			logger.Error("replace filter config rule count type is invalid")
			return replaceRule{}, false
		}
	}
	return replaceRule{
		getter: field_getter.GetFieldGetter(name),
		setter: field_setter.NewFieldSetter(name, true),
		old:    old,
		new:    newStr,
		count:  count,
	}, true
}

// 若遇到一个字段处理失败，则继续处理其它的，最后返回错误
func (f *ReplaceFilter) Filter(event *event.Event) (*event.Event, error) {
	var failed bool
	for _, rule := range f.fields {
		value, err := rule.getter.GetField(event)
		if err != nil {
			// 字段不存在则跳过，不做任何操作
			if !errors.Is(err, field_getter.ErrNotFound) {
				failed = true
			}
			continue
		}
		v, ok := value.(string)
		if !ok {
			failed = true
			continue
		}
		rule.setter.SetField(event, strings.Replace(v, rule.old, rule.new, rule.count))
	}
	if failed {
		return event, errors.New("replace filter failed")
	}
	return event, nil
}
