package filter

import (
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

func TestAdd(t *testing.T) {
	conf := map[any]any{
		"overwrite": true,
		"fields": map[any]any{
			"flag":  "test",
			"count": 1,
		},
	}
	f := newAddFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newAddFilter")
	}

	e := &event.Event{Data: map[string]any{"level": "DEBUG"}}
	if _, err := f.Filter(e); err != nil {
		t.Fatalf("add filter failed: %v", err)
	}
	if e.Data["flag"] != "test" {
		t.Errorf("add filter failed: flag = %v, expected test", e.Data["flag"])
	}
	if e.Data["count"] != 1 {
		t.Errorf("add filter failed: count = %v, expected 1", e.Data["count"])
	}
}

func TestAddOverwriteFalse(t *testing.T) {
	conf := map[any]any{
		"overwrite": false,
		"fields": map[any]any{
			"level": "INFO",
		},
	}
	f := newAddFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newAddFilter")
	}

	// 已存在的字段不覆盖
	e := &event.Event{Data: map[string]any{"level": "DEBUG"}}
	if _, err := f.Filter(e); err != nil {
		t.Fatalf("add filter failed: %v", err)
	}
	if e.Data["level"] != "DEBUG" {
		t.Errorf("add filter failed: level = %v, expected DEBUG (not overwritten)", e.Data["level"])
	}

	// 不存在的字段正常添加
	e = &event.Event{Data: map[string]any{}}
	if _, err := f.Filter(e); err != nil {
		t.Fatalf("add filter failed: %v", err)
	}
	if e.Data["level"] != "INFO" {
		t.Errorf("add filter failed: level = %v, expected INFO", e.Data["level"])
	}
}

func TestAddMultiLevel(t *testing.T) {
	conf := map[any]any{
		"fields": map[any]any{
			"[name][first]": "John",
		},
	}
	f := newAddFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newAddFilter")
	}

	e := &event.Event{Data: map[string]any{}}
	if _, err := f.Filter(e); err != nil {
		t.Fatalf("add filter failed: %v", err)
	}
	name, ok := e.Data["name"].(map[string]any)
	if !ok {
		t.Fatalf("add filter failed: name = %v, expected map", e.Data["name"])
	}
	if name["first"] != "John" {
		t.Errorf("add filter failed: name.first = %v, expected John", name["first"])
	}
}
