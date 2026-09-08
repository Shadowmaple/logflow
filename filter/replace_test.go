package filter

import (
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

func TestReplaceAll(t *testing.T) {
	conf := map[any]any{
		"fields": map[any]any{
			"msg": []any{"en", "eng"},
		},
	}
	f := newReplaceFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newReplaceFilter")
	}

	e := &event.Event{Data: map[string]any{"msg": "en en x"}}
	if _, err := f.Filter(e); err != nil {
		t.Fatalf("replace filter failed: %v", err)
	}
	if e.Data["msg"] != "eng eng x" {
		t.Errorf("replace filter failed: msg = %v, expected 'eng eng x'", e.Data["msg"])
	}
}

func TestReplaceOnce(t *testing.T) {
	conf := map[any]any{
		"fields": map[any]any{
			"msg": []any{"wang", "Wang", 1},
		},
	}
	f := newReplaceFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newReplaceFilter")
	}

	e := &event.Event{Data: map[string]any{"msg": "wang wang"}}
	if _, err := f.Filter(e); err != nil {
		t.Fatalf("replace filter failed: %v", err)
	}
	if e.Data["msg"] != "Wang wang" {
		t.Errorf("replace filter failed: msg = %v, expected 'Wang wang'", e.Data["msg"])
	}
}

func TestReplaceFieldNotFound(t *testing.T) {
	conf := map[any]any{
		"fields": map[any]any{
			"msg": []any{"en", "eng"},
		},
	}
	f := newReplaceFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newReplaceFilter")
	}

	// 字段不存在时不做任何操作，不返回错误
	e := &event.Event{Data: map[string]any{"level": "DEBUG"}}
	if _, err := f.Filter(e); err != nil {
		t.Errorf("replace filter failed: expected nil error when field not found, got %v", err)
	}
}

func TestReplaceNotString(t *testing.T) {
	conf := map[any]any{
		"fields": map[any]any{
			"count": []any{"1", "one"},
		},
	}
	f := newReplaceFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newReplaceFilter")
	}

	// 字段值不是字符串时返回错误
	e := &event.Event{Data: map[string]any{"count": 1}}
	if _, err := f.Filter(e); err == nil {
		t.Error("replace filter failed: expected error when value is not a string")
	}
}

func TestReplaceMultiLevel(t *testing.T) {
	conf := map[any]any{
		"fields": map[any]any{
			"[name][last]": []any{"smith", "lee"},
		},
	}
	f := newReplaceFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newReplaceFilter")
	}

	e := &event.Event{Data: map[string]any{
		"name": map[string]any{"last": "smith smith"},
	}}
	if _, err := f.Filter(e); err != nil {
		t.Fatalf("replace filter failed: %v", err)
	}
	name := e.Data["name"].(map[string]any)
	if name["last"] != "lee lee" {
		t.Errorf("replace filter failed: name.last = %v, expected 'lee lee'", name["last"])
	}
}
