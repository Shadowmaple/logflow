package utils

import (
	"fmt"
	"strings"
)

// 将map转为json字符串
func MapToJsonString(m map[string]any) string {
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
// var bufPool = sync.Pool{
//     New: func() interface{} {
//         return new(bytes.Buffer)
//     },
// }

// func MapToJSON(m map[string]interface{}) ([]byte, error) {
//     buf := bufPool.Get().(*bytes.Buffer)
//     buf.Reset()
//     defer bufPool.Put(buf)

//     if err := json.NewEncoder(buf).Encode(m); err != nil {
//         return nil, err
//     }

//     // Encode 会添加换行符 \n，如不需要可 trim
//     jsonBytes := bytes.TrimRight(buf.Bytes(), "\n")
//     return jsonBytes, nil
// }
