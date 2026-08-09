package render

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Shadowmaple/logflow/internal/event"
)

type field struct {
	literal bool // 是否为字面量，即不需要赋值
	value   string
	mu      *MultiLevelRender // 多层级解析器，如[@metadata][kafka][topic]
}

func (f *field) render(event *event.Event) (string, error) {
	if f.literal {
		return f.value, nil
	}
	if f.mu != nil {
		val, err := f.mu.Render(event)
		if err != nil {
			return "", err
		}
		return val.(string), nil
	}
	if val, ok := event.Data[f.value]; ok {
		return val.(string), nil
	}
	return "", fmt.Errorf("index render failed, %s not found", f.value)
}

func newIndexRender(template string) *IndexRender {
	r, _ := regexp.Compile(`%{(.+?)}`)
	fields := make([]*field, 0)
	lastIdx := 0
	for _, v := range r.FindAllStringIndex(template, -1) {
		l, r := v[0], v[1]
		// 不需要赋值的字段
		fields = append(fields, &field{
			literal: true,
			value:   template[lastIdx:l],
		})
		// 需赋值的字段
		// TODO: 多级嵌套的字段，如[@metadata][kafka][topic]
		val := template[l+2 : r-1]
		if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
			fields = append(fields, &field{
				literal: false,
				value:   val,
				mu:      newMultiLevelRender(getAllFields(val)),
			})
		} else {
			fields = append(fields, &field{
				literal: false,
				value:   val,
			})
		}
		lastIdx = r
	}
	if lastIdx < len(template) {
		fields = append(fields, &field{
			literal: true,
			value:   template[lastIdx:],
		})
	}
	return &IndexRender{fields: fields}
}

// getAllFields ("%{[@metadata][kafka][topic]}") => ["@metadata","kafka","topic"]
func getAllFields(s string) []string {
	fields := make([]string, 0)
	r, _ := regexp.Compile(`\[(.*?)\]`)
	for _, v := range r.FindAll([]byte(s), -1) {
		fields = append(fields, string(v[1:len(v)-1]))
	}
	return fields
}

type IndexRender struct {
	fields []*field
}

func (r *IndexRender) Render(event *event.Event) (any, error) {
	values := make([]string, 0, len(r.fields))
	for _, f := range r.fields {
		val, err := f.render(event)
		if err != nil {
			return "", err
		}
		values = append(values, val)
	}
	return strings.Join(values, ""), nil
}
