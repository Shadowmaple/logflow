package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 递归将 map[string]any 类型转换为 json 形式的 map[string]any
func ConvertToJSONCompatible(conf any) any {
	switch v := conf.(type) {
	case map[any]any:
		result := make(map[string]any)
		for k, val := range v {
			if keyStr, ok := k.(string); ok {
				result[keyStr] = ConvertToJSONCompatible(val)
			} else {
				panic(fmt.Sprintf("config key '%v' is not a string", k))
			}
		}
		return result
	case map[string]any:
		result := make(map[string]any)
		for k, val := range v {
			result[k] = ConvertToJSONCompatible(val)
		}
		return result
	case []any:
		result := make([]any, len(v))
		for i, val := range v {
			result[i] = ConvertToJSONCompatible(val)
		}
		return result
	default:
		return v
	}
}

// 将配置文件通过 json 标准库解析到 result 中
func SafeDecodeConfig(kind string, config map[string]any, result any) {
	// Convert config to JSON-serializable format
	jsonConfig := ConvertToJSONCompatible(config)

	// Convert map to JSON and then unmarshal to struct
	jsonBytes, err := json.Marshal(jsonConfig)
	if err != nil {
		panic(fmt.Sprintf("%s type: failed to marshal config to JSON: %v", kind, err))
	}

	// Use a decoder with UseNumber to preserve number precision and allow type flexibility
	decoder := json.NewDecoder(strings.NewReader(string(jsonBytes)))
	decoder.UseNumber()

	if err := decoder.Decode(result); err != nil {
		panic(fmt.Sprintf("%s type configuration error: %v", kind, err))
	}
}

// 校验配置文件中是否缺少必填字段
func ValidateRequiredFields(kind string, fields map[string]any) {
	for fieldName, fieldValue := range fields {
		if fieldValue == nil || (fmt.Sprintf("%v", fieldValue) == "") {
			panic(fmt.Sprintf("%s type: '%s' is required", kind, fieldName))
		}
	}
}
