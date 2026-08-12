package field_setter

import (
	"reflect"
	"testing"
	"time"
)

func TestOneLevelFieldSetter_WritesWhenAbsent(t *testing.T) {
	f := newOneLevelFieldSetter("message", true)
	ev := newFieldEvent(nil)

	f.SetField(ev, "hello")
	if got := ev.Data["message"]; got != "hello" {
		t.Fatalf("message = %v, want %q", got, "hello")
	}
}

func TestOneLevelFieldSetter_OverwriteTrue_OverwritesExisting(t *testing.T) {
	f := newOneLevelFieldSetter("message", true)
	ev := newFieldEvent(map[string]any{"message": "old"})

	f.SetField(ev, "new")
	if got := ev.Data["message"]; got != "new" {
		t.Fatalf("message = %v, want %q (should be overwritten)", got, "new")
	}
}

func TestOneLevelFieldSetter_OverwriteFalse_SkipsExisting(t *testing.T) {
	f := newOneLevelFieldSetter("message", false)
	ev := newFieldEvent(map[string]any{"message": "old"})

	f.SetField(ev, "new")
	if got := ev.Data["message"]; got != "old" {
		t.Fatalf("message = %v, want %q (should not be overwritten)", got, "old")
	}
}

func TestOneLevelFieldSetter_OverwriteFalse_WritesWhenAbsent(t *testing.T) {
	// overwrite=false 但字段不存在时仍应写入
	f := newOneLevelFieldSetter("message", false)
	ev := newFieldEvent(nil)

	f.SetField(ev, "hello")
	if got := ev.Data["message"]; got != "hello" {
		t.Fatalf("message = %v, want %q", got, "hello")
	}
}

func TestOneLevelFieldSetter_WritesVariousTypes(t *testing.T) {
	// 验证 SetField 可写入任意类型的值
	now := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	cases := []struct {
		name  string
		value any
	}{
		{"string", "hello"},
		{"int", 42},
		{"bool", true},
		{"time", now},
		{"slice", []any{1, 2, 3}},
		{"map", map[string]any{"k": "v"}},
		{"nil", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newOneLevelFieldSetter("field", true)
			ev := newFieldEvent(nil)
			f.SetField(ev, c.value)
			if !reflect.DeepEqual(ev.Data["field"], c.value) {
				t.Fatalf("field = %v, want %v", ev.Data["field"], c.value)
			}
		})
	}
}
