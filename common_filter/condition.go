package commonFilter

import (
	"reflect"

	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/model"

	"go.uber.org/zap"
)

// if 条件顾虑器
type ConditionFilter struct {
	conditions []*conditionNode
}

func NewConditionFilter(conf map[string]any) *ConditionFilter {
	if conf == nil {
		logger.Warn("condition filter config is nil")
		return nil
	}
	entry, ok := conf["if"]
	if !ok {
		logger.Warn("condition filter config is missing if expression")
		return nil
	}

	// 判断if条件配置类型，将每个条件表达式解析为条件判断树
	var expressions []*conditionNode
	switch value := entry.(type) {
	case []any:
		expressions = make([]*conditionNode, 0, len(value))
		for _, v := range value {
			if raw, ok := v.(string); ok {
				if expr, ok := parseConditionExpr(raw); ok {
					expressions = append(expressions, expr)
				} else {
					logger.Error("failed to parse condition expression", zap.String("value", raw))
				}
			} else {
				logger.Error("unsupported condition filter config type in array", zap.String("type", reflect.TypeOf(v).String()))
			}
		}
	case []string:
		expressions = make([]*conditionNode, 0, len(value))
		for _, raw := range value {
			if expr, ok := parseConditionExpr(raw); ok {
				expressions = append(expressions, expr)
			} else {
				logger.Error("failed to parse condition expression", zap.String("value", raw))
			}
		}
	case string:
		if expr, ok := parseConditionExpr(value); ok {
			expressions = append(expressions, expr)
		} else {
			logger.Error("failed to parse condition expression", zap.String("value", value))
		}
	default:
		logger.Error("unsupported condition filter config type", zap.String("type", reflect.TypeOf(entry).String()))
		return nil
	}

	if len(expressions) == 0 {
		logger.Warn("condition filter has no valid expressions")
		return nil
	}
	return &ConditionFilter{conditions: expressions}
}

// 是否通过判断条件
func (c *ConditionFilter) Check(event *model.Event) bool {
	if c == nil || len(c.conditions) == 0 {
		return true
	}
	if event == nil {
		return false
	}
	if event.Data == nil {
		logger.Warn("condition filter received event without data map")
		return false
	}
	for _, expr := range c.conditions {
		if !expr.check(event.Data) {
			return false
		}
	}
	return true
}

type conditionNode struct {
	op     uint8
	left   *conditionNode
	right  *conditionNode
	clause *conditionClause
}

// 条件是否成立
func (e *conditionNode) check(data map[string]any) bool {
	if e == nil {
		return false
	}

	switch e.op {
	case _op_and:
		if e.left == nil || e.right == nil {
			return false
		}
		return e.left.check(data) && e.right.check(data)
	case _op_or:
		if e.left == nil || e.right == nil {
			return false
		}
		return e.left.check(data) || e.right.check(data)
	default:
		if e.clause == nil {
			return true
		}
		return e.clause.matches(data)
	}
}
