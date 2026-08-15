package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// 将map转为json字符串，只支持顶层转化
func MapToJsonString(m map[any]any) string {
	var jsonStr strings.Builder
	for k, v := range m {
		fmt.Fprintf(&jsonStr, "%s: %v, ", k, v)
	}
	return jsonStr.String()
}

// 字符串清理：去除首尾空格、换行符、制表符
func TrimStr(s string) string {
	return strings.Trim(s, " \n\t")
}

// 使用 sync.Pool 复用 bytes.Buffer
// 避免 Marshal 的额外分配
var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// 将map转为json字节序列，使用json序列化
func MapToJSON(m map[any]any) ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	data := ConvertToJSONCompatible(m)

	if err := json.NewEncoder(buf).Encode(data); err != nil {
		return nil, err
	}

	// Encode 会添加换行符 \n，如不需要可 trim
	jsonBytes := bytes.TrimRight(buf.Bytes(), "\n")
	return jsonBytes, nil
}

// // 将map转为json字节序列，使用json序列化，格式化输出
// func MapToPrettyJSON(m map[any]any) ([]byte, error) {
// 	buf := bufPool.Get().(*bytes.Buffer)
// 	buf.Reset()
// 	defer bufPool.Put(buf)

// 	if err := json.NewEncoder(buf).Encode(m); err != nil {
// 		return nil, err
// 	}

// 	// Encode 会添加换行符 \n，如不需要可 trim
// 	jsonBytes := bytes.TrimRight(buf.Bytes(), "\n")
// 	return jsonBytes, nil
// }
