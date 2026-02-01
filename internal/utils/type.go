package utils

import (
	"strconv"
	"strings"
)

// 判断值的类型
func TypeCheck(v any, t string) bool {
	switch t {
	case "string":
		_, ok := v.(string)
		return ok
	case "int":
		_, ok := v.(int)
		return ok
	case "float":
		_, ok := v.(float64)
		return ok
	case "bool":
		_, ok := v.(bool)
		return ok
	default:
		return false
	}
}

// 判断值类型，转为bool
// 返回值：
//  1. 转换后的bool值
//  2. 是否转换成功
func ParseToBool(value any) (bool, bool) {
	switch v := value.(type) {
	case string:
		v = strings.ToLower(strings.Trim(v, ""))
		switch v {
		case "true", "on":
			return true, true
		case "false", "off":
			return false, true
		default:
			return false, false
		}
	case bool:
		return v, true
	case int:
		return v != 0, true
	}
	return false, false
}

func ParseToStrList(value any) ([]string, bool) {
	switch v := value.(type) {
	case string:
		strs := strings.Split(TrimStr(v), ",")
		list := make([]string, len(strs))
		for i, str := range strs {
			list[i] = TrimStr(str)
		}
		return list, true
	case []string:
		return v, true
	}
	return nil, false
}

func ParseToInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case string:
		var i int
		i, err := strconv.Atoi(TrimStr(v))
		if err != nil {
			return 0, false
		}
		return i, true
	case bool:
		if v {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}
