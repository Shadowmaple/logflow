package filter

import (
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

func TestLowercaseFilter_SingleField(t *testing.T) {
	f := newLowercaseFilter(map[string]any{
		"fields": []string{"message"},
	})
	ev := &event.Event{Data: map[string]any{"message": "Hello WORLD 123"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["message"]; got != "hello world 123" {
		t.Fatalf("message = %v, want %q", got, "hello world 123")
	}
}

func TestLowercaseFilter_MultipleFields(t *testing.T) {
	f := newLowercaseFilter(map[string]any{
		"fields": []string{"message", "host"},
	})
	ev := &event.Event{Data: map[string]any{
		"message": "Hello",
		"host":    "LOCALHOST",
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["message"]; got != "hello" {
		t.Fatalf("message = %v, want %q", got, "hello")
	}
	if got := ev.Data["host"]; got != "localhost" {
		t.Fatalf("host = %v, want %q", got, "localhost")
	}
}

func TestLowercaseFilter_AlreadyLowercase(t *testing.T) {
	f := newLowercaseFilter(map[string]any{
		"fields": []string{"message"},
	})
	ev := &event.Event{Data: map[string]any{"message": "already lower"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["message"]; got != "already lower" {
		t.Fatalf("message = %v, want %q", got, "already lower")
	}
}

func TestLowercaseFilter_BracketSingleLevel(t *testing.T) {
	// "[message]" 等价于单层级字段 message
	f := newLowercaseFilter(map[string]any{
		"fields": []string{"[message]"},
	})
	ev := &event.Event{Data: map[string]any{"message": "HELLO"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["message"]; got != "hello" {
		t.Fatalf("message = %v, want %q", got, "hello")
	}
}

func TestLowercaseFilter_EmptyFields(t *testing.T) {
	f := newLowercaseFilter(map[string]any{
		"fields": []string{},
	})
	ev := &event.Event{Data: map[string]any{"message": "HELLO"}}

	_, err := f.Filter(ev)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 没有字段需要处理，事件保持不变
	if got := ev.Data["message"]; got != "HELLO" {
		t.Fatalf("message = %v, want %q", got, "HELLO")
	}
}

func TestLowercaseFilter_FieldNotFound(t *testing.T) {
	f := newLowercaseFilter(map[string]any{
		"fields": []string{"missing"},
	})
	ev := &event.Event{Data: map[string]any{"message": "HELLO"}}

	_, err := f.Filter(ev)
	if err == nil {
		t.Fatal("expected error for missing field, got nil")
	}
	if err.Error() != "lowercase filter failed" {
		t.Fatalf("error = %q, want %q", err.Error(), "lowercase filter failed")
	}
	// 不存在的字段不应被写入
	if _, ok := ev.Data["missing"]; ok {
		t.Fatal("missing field should not be created")
	}
}

func TestLowercaseFilter_NonStringValue(t *testing.T) {
	f := newLowercaseFilter(map[string]any{
		"fields": []string{"count"},
	})
	ev := &event.Event{Data: map[string]any{"count": 123}}

	_, err := f.Filter(ev)
	if err == nil {
		t.Fatal("expected error for non-string value, got nil")
	}
	// 非字符串值保持不变
	if got := ev.Data["count"]; got != 123 {
		t.Fatalf("count = %v, want 123", got)
	}
}

func TestLowercaseFilter_MixedSuccessAndFailure(t *testing.T) {
	// 部分字段处理成功、部分失败时，成功的字段仍应被转换，
	// 同时整体返回错误
	f := newLowercaseFilter(map[string]any{
		"fields": []string{"message", "missing"},
	})
	ev := &event.Event{Data: map[string]any{"message": "HELLO"}}

	_, err := f.Filter(ev)
	if err == nil {
		t.Fatal("expected error due to missing field, got nil")
	}
	if got := ev.Data["message"]; got != "hello" {
		t.Fatalf("message = %v, want %q (should still be lowercased)", got, "hello")
	}
}

func TestNewLowercaseFilter_PanicsOnMissingFields(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when fields missing, got none")
		}
	}()
	newLowercaseFilter(map[string]any{})
}

func TestNewLowercaseFilter_PanicsOnWrongType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when fields is wrong type, got none")
		}
	}()
	// []any 不满足 []string 断言
	newLowercaseFilter(map[string]any{
		"fields": []any{"message"},
	})
}

func TestNewLowercaseFilter_PanicsOnNilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when config is nil, got none")
		}
	}()
	newLowercaseFilter(nil)
}

func TestBuildFilter_Lowercase(t *testing.T) {
	// 通过注册名 "lowercase" 构建 filter
	conf := map[string]any{
		"lowercase": map[string]any{
			"fields": []string{"message"},
		},
	}
	f := BuildFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from BuildFilter")
	}
	ev := &event.Event{Data: map[string]any{"message": "BuildFilter TEST"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["message"]; got != "buildfilter test" {
		t.Fatalf("message = %v, want %q", got, "buildfilter test")
	}
}

func TestBuildFilter_LowercaseUnknownType(t *testing.T) {
	// 未注册的 filter 类型应返回 nil
	f := BuildFilter(map[string]any{
		"not_a_real_filter": map[string]any{},
	})
	if f != nil {
		t.Fatalf("expected nil filter for unknown type, got %T", f)
	}
}
