package render

import (
	"regexp"

	"github.com/Shadowmaple/logflow/internal/model"
)

type field struct {
	literal bool // 是否为字面量，即不需要赋值
	value   string
}

func (f *field) render(data map[string]any) string {
	if f.literal {
		return f.value
	}
	if val, ok := data[f.value]; ok {
		return val.(string)
	}
	return ""
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
		fields = append(fields, &field{
			literal: false,
			value:   template[l+2 : r-1],
		})
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

type IndexRender struct {
	fields []*field
}

func (r *IndexRender) Render(event *model.Event) (any, error) {
	data := event.Data
	result := ""
	for _, f := range r.fields {
		result += f.render(data)
	}
	return result, nil
}
