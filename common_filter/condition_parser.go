package commonFilter

import (
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/Shadowmaple/logflow/internal/logger"
	"go.uber.org/zap"
)

const (
	_op_omit = iota
	_op_or
	_op_and
	_op_not

	_op_exist = iota + 32
	_op_eq
	_op_in
	_op_notin

	// 表达式解析状态
	_state_default   = iota // 无操作
	_state_condition        // 解析表达式子串中
	_state_operator         // 解析运算符中
)

type conditionClause struct {
	op     uint8
	field  string
	expect any
}

// 解析条件表达式
func parseConditionExpr(raw string) (*conditionNode, bool) {
	expr, err := parseToConditionNode(strings.TrimSpace(raw))
	if err != nil {
		logger.Error("failed to parse condition expression", zap.String("raw", raw), zap.Error(err))
		return nil, false
	}
	return expr, true
}

// 将条件表达式解析为条件判断树
func parseToConditionNode(raw string) (node *conditionNode, err error) {
	defer func() {
		if r := recover(); r != nil {
			node = nil
			err = fmt.Errorf("parse `%s` error at `%s`", raw, r)
		}
	}()

	// 将中缀表达式转为逆波兰式
	rpnItems, err := buildRPN(raw)
	if err != nil {
		return
	}

	// 将逆波兰式表达式构建为条件表达式树
	stack := []*conditionNode{}
	for _, item := range rpnItems {
		switch v := item.(type) {
		case string:
			clause, ok := parseClause(v)
			if !ok {
				return nil, fmt.Errorf("failed to parse clause: %s", v)
			}
			stack = append(stack, &conditionNode{clause: &clause})
		case uint8: // 操作符
			if len(stack) < 1 {
				return nil, fmt.Errorf("not enough operands for operator: %d", v)
			}
			if v == _op_not {
				right := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				stack = append(stack, &conditionNode{op: v, right: right})
			} else {
				if len(stack) < 2 {
					return nil, fmt.Errorf("not enough operands for operator: %d", v)
				}
				right := stack[len(stack)-1]
				left := stack[len(stack)-2]
				stack = stack[:len(stack)-2]
				stack = append(stack, &conditionNode{op: v, left: left, right: right})
			}
		default:
			return nil, fmt.Errorf("unexpected RPN item type: %T", v)
		}
	}

	if len(stack) != 1 {
		return nil, fmt.Errorf("invalid expression, stack size: %d", len(stack))
	}

	return stack[0], nil
}

// 将条件表达式由中缀表达式转为逆波兰式（后缀表达式）
// 将)作为表达式方法的结束标识
func buildRPN(raw string) ([]any, error) {
	res := []any{}
	condition_start := 0
	var state = _state_default
	ops := []uint8{} // 操作符栈

	for i := 0; i < len(raw); i++ {
		if raw[i] == ' ' || raw[i] == '\t' || raw[i] == '\n' || raw[i] == '\r' {
			continue
		}
		switch raw[i] {
		case '|':
			switch state {
			case _state_default:
				state = _state_operator
				if i+1 >= len(raw) || raw[i+1] != '|' {
					return nil, fmt.Errorf("unexpected character '|' at position %d", i)
				}
				pushOp(_op_or, &res, &ops)
				i++
			}
		case '&':
			switch state {
			case _state_default:
				state = _state_operator
				if i+1 >= len(raw) || raw[i+1] != '&' {
					return nil, fmt.Errorf("unexpected character '&' at position %d", i)
				}
				pushOp(_op_and, &res, &ops)
				i++
			}
		// case '(':
		// 	stack = append(stack, "(")
		case ')':
			switch state {
			case _state_condition:
				// 正则解析表达式中，) 为表达式的一部分，则将表达式子句入栈
				res = append(res, raw[condition_start:i+1])
				state = _state_default
			}
		default:
			if state != _state_condition {
				condition_start = i
				state = _state_condition
			}
		}
	}
	// 将剩余操作符入栈
	for _, op := range slices.Backward(ops) {
		res = append(res, op)
	}

	return res, nil
}

// 将栈加入栈结果中
// res: 输出的结果栈
// ops: 当前运算符栈
func pushOp(op uint8, res *[]any, ops *[]uint8) {
	// 若ops有运算符，则比较优先级，若ops栈顶运算符优先级大于等于当前运算符，则将ops栈顶运算符出栈并入栈到结果中
	for len(*ops) > 0 {
		top := (*ops)[len(*ops)-1]
		if top >= op {
			// 栈顶运算符优先级大于等于当前运算符，出栈并入栈到结果中
			*res = append(*res, top)
			*ops = (*ops)[:len(*ops)-1]
		} else {
			break
		}
	}
	// 将当前运算符入栈
	*ops = append(*ops, op)
}

// 解析单个条件判定方法，如EQ(field, "value")、EXIST(field)等
func parseClause(raw string) (conditionClause, bool) {
	left := strings.Index(raw, "(")
	right := strings.LastIndex(raw, ")")
	if left <= 0 || right <= left+1 || right != len(raw)-1 {
		logger.Error("不支持的条件格式", zap.String("raw", raw))
		return conditionClause{}, false
	}

	name := strings.ToUpper(strings.TrimSpace(raw[:left]))
	args := splitArgs(strings.TrimSpace(raw[left+1 : right]))
	logger.Debug(fmt.Sprintf("parseClause name=%s, args=%v", name, args))

	switch name {
	case "EXIST":
		if len(args) != 1 {
			logger.Error("EXIST operator requires exactly 1 argument", zap.String("raw", raw))
			return conditionClause{}, false
		}
		return conditionClause{op: _op_exist, field: normalizeIdentifier(args[0])}, true
	case "EQ":
		if len(args) != 2 {
			logger.Error("EQ operator requires exactly 2 arguments", zap.String("raw", raw))
			return conditionClause{}, false
		}
		return conditionClause{op: _op_eq, field: normalizeIdentifier(args[0]), expect: parseLiteral(args[1])}, true
	case "IN":
		if len(args) != 2 {
			logger.Error("IN operator requires exactly 2 arguments", zap.String("raw", raw))
			return conditionClause{}, false
		}
		return conditionClause{op: _op_in, field: normalizeIdentifier(args[0]), expect: parseLiteral(args[1])}, true
	case "NOTIN":
		if len(args) != 2 {
			logger.Error("NOTIN operator requires exactly 2 arguments", zap.String("raw", raw))
			return conditionClause{}, false
		}
		return conditionClause{op: _op_notin, field: normalizeIdentifier(args[0]), expect: parseLiteral(args[1])}, true
	default:
		logger.Error("不支持的条件方法:"+name, zap.String("raw", raw))
		return conditionClause{}, false
	}
}

// 拆分条件方法参数
func splitArgs(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// 处理字段字面量，去除空白符、引号等，返回原始值
func normalizeIdentifier(raw string) string {
	trimmed := strings.TrimSpace(raw)
	// 去除引号
	if len(trimmed) >= 2 {
		if (trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"') || (trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'') {
			return trimmed[1 : len(trimmed)-1]
		}
	}
	return trimmed
}

// 解析值与类型
// 字符串、数字、nil等
func parseLiteral(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	// 字符串类型
	if len(trimmed) >= 2 {
		if (trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"') || (trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'') {
			return trimmed[1 : len(trimmed)-1]
		}
	}
	// nil 值
	if strings.EqualFold(trimmed, "nil") {
		return nil
	}
	// 整型
	if value, err := strconv.Atoi(trimmed); err == nil {
		return value
	}
	logger.Warn("unsupported literal type, treating as string", zap.String("value", trimmed))
	return trimmed
}

// 每个条件子句的匹配是否成立
func (c *conditionClause) matches(data map[string]any) bool {
	value, ok := data[c.field]
	if !ok {
		return false
	}

	switch c.op {
	case _op_exist:
		return true
	case _op_eq:
		return compareValues(value, c.expect)
	case _op_in:
		return containsValue(value, c.expect)
	case _op_notin:
		return !containsValue(value, c.expect)
	default:
		return false
	}
}

func compareValues(actual, expected any) bool {
	if expected == nil {
		return actual == nil
	}
	if actual == nil {
		return false
	}

	switch expectedValue := expected.(type) {
	case string:
		actualValue, ok := actual.(string)
		return ok && actualValue == expectedValue
	case int:
		switch actualValue := actual.(type) {
		case int:
			return actualValue == expectedValue
		case int8:
			return int(actualValue) == expectedValue
		case int16:
			return int(actualValue) == expectedValue
		case int32:
			return int(actualValue) == expectedValue
		case int64:
			return int(actualValue) == expectedValue
		case uint:
			return int(actualValue) == expectedValue
		case uint8:
			return int(actualValue) == expectedValue
		case uint16:
			return int(actualValue) == expectedValue
		case uint32:
			return int(actualValue) == expectedValue
		case uint64:
			return int(actualValue) == expectedValue
		// case float32:
		// 	return int(actualValue) == expectedValue
		// case float64:
		// 	return int(actualValue) == expectedValue
		default:
			return false
		}
	default:
		return false
	}
}

func containsValue(actual, expected any) bool {
	if actual == nil {
		return false
	}

	value := reflect.ValueOf(actual)
	if value.Kind() != reflect.Slice && value.Kind() != reflect.Array {
		return false
	}

	for i := 0; i < value.Len(); i++ {
		if compareValues(value.Index(i).Interface(), expected) {
			return true
		}
	}
	return false
}
