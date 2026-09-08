package field_setter

import "testing"

func TestMultiLevelFieldSetter_TwoLevels(t *testing.T) {
	f := newMultiLevelFieldSetter([]string{"a", "b"}, true)
	ev := newFieldEvent(nil)

	f.SetField(ev, "hello")

	got, ok := ev.Data["a"].(map[string]any)
	if !ok {
		t.Fatalf("a should be map[string]any, got %T", ev.Data["a"])
	}
	if got["b"] != "hello" {
		t.Fatalf(`a[b] = %v, want %q`, got["b"], "hello")
	}
}

func TestMultiLevelFieldSetter_NonMapIntermediate_Skips(t *testing.T) {
	// 中间层字段已存在但非 map 时，无法继续下钻，应跳过且不修改原值
	f := newMultiLevelFieldSetter([]string{"a", "b"}, true)
	ev := newFieldEvent(map[string]any{"a": "not-a-map"})

	f.SetField(ev, "hello")

	// 数据保持不变
	assertDataEqual(t, ev, map[string]any{"a": "not-a-map"})
}

func TestMultiLevelFieldSetter_OverwriteFalse_SkipsLastField(t *testing.T) {
	f := newMultiLevelFieldSetter([]string{"a", "b"}, false)
	ev := newFieldEvent(map[string]any{
		"a": map[string]any{
			"b": map[string]any{"b": "old"},
		},
	})

	f.SetField(ev, "new")

	// lastField "b" 在 Data["a"]["b"] 中已存在且 overwrite=false -> 不覆盖
	got := ev.Data["a"].(map[string]any)["b"].(map[string]any)["b"]
	if got != "old" {
		t.Fatalf(`a[b][b] = %v, want %q (overwrite=false should skip)`, got, "old")
	}
}

func TestMultiLevelFieldSetter_OverwriteTrue_OverwritesLastField(t *testing.T) {
	// overwrite=true 时覆盖 lastField 位置
	f := newMultiLevelFieldSetter([]string{"a", "b"}, true)
	ev := newFieldEvent(map[string]any{
		"a": map[string]any{
			"b": map[string]any{"b": "old"},
		},
	})

	f.SetField(ev, "new")

	got := ev.Data["a"].(map[string]any)["b"]
	if got != "new" {
		t.Fatalf(`a[b] = %v, want %q (overwrite=true should overwrite)`, got, "new")
	}
}

func TestNewFieldSetter_MultipleBrackets_CreatesIntermediateMaps(t *testing.T) {
	// 通过 NewFieldSetter 构建 MultiLevelFieldSetter，验证中间层 map 被创建
	f := NewFieldSetter("[a][b]", true)
	ev := newFieldEvent(nil)

	f.SetField(ev, "hello")

	want := map[string]any{
		"a": map[string]any{
			"b": "hello",
		},
	}
	assertDataEqual(t, ev, want)
}
