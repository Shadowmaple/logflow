package filter

import (
	"reflect"
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

func TestJsonFilter_WithTarget(t *testing.T) {
	// 解析结果应存储到 target 字段，source 保持不变
	f := newJsonFilter(map[any]any{
		"source": "raw",
		"target": "parsed",
	})
	ev := &event.Event{Data: map[string]any{
		"raw": []byte(`{"message":"hello","count":5}`),
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// source 仍保留为 []byte
	if _, ok := ev.Data["raw"].([]byte); !ok {
		t.Fatalf("raw should remain []byte, got %T", ev.Data["raw"])
	}
	got, ok := ev.Data["parsed"].(map[string]any)
	if !ok {
		t.Fatalf("parsed should be map[string]any, got %T", ev.Data["parsed"])
	}
	if got["message"] != "hello" {
		t.Fatalf("message = %v, want %q", got["message"], "hello")
	}
	if got["count"] != float64(5) {
		t.Fatalf("count = %v, want 5", got["count"])
	}
}

func TestJsonFilter_WithEmptyTarget_MergesIntoData(t *testing.T) {
	// target 为空时，解析结果应通过 maps.Copy 合并进 event.Data
	f := newJsonFilter(map[any]any{
		"source": "raw",
	})
	ev := &event.Event{Data: map[string]any{
		"raw":      []byte(`{"message":"hello","count":5}`),
		"existing": "keep",
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["message"]; got != "hello" {
		t.Fatalf("message = %v, want %q", got, "hello")
	}
	if got := ev.Data["count"]; got != float64(5) {
		t.Fatalf("count = %v, want 5", got)
	}
	// 已有字段与 source 应保留
	if got := ev.Data["existing"]; got != "keep" {
		t.Fatalf("existing = %v, want %q", got, "keep")
	}
	if _, ok := ev.Data["raw"].([]byte); !ok {
		t.Fatalf("raw should remain []byte, got %T", ev.Data["raw"])
	}
}

func TestJsonFilter_EmptyTarget_OverwritesExistingKeys(t *testing.T) {
	// 合并时，解析结果中与 event.Data 同名的键应覆盖原值
	f := newJsonFilter(map[any]any{
		"source": "raw",
	})
	ev := &event.Event{Data: map[string]any{
		"raw":     []byte(`{"message":"new"}`),
		"message": "old",
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["message"]; got != "new" {
		t.Fatalf("message = %v, want %q (should be overwritten)", got, "new")
	}
}

func TestJsonFilter_EmptyJsonObject(t *testing.T) {
	// 空对象 {} 是合法 JSON，解析后为空 map
	f := newJsonFilter(map[any]any{
		"source": "raw",
		"target": "parsed",
	})
	ev := &event.Event{Data: map[string]any{
		"raw": []byte(`{}`),
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := ev.Data["parsed"].(map[string]any)
	if !ok {
		t.Fatalf("parsed should be map[string]any, got %T", ev.Data["parsed"])
	}
	if len(got) != 0 {
		t.Fatalf("parsed should be empty map, got %v", got)
	}
}

func TestJsonFilter_SourceNotFound(t *testing.T) {
	f := newJsonFilter(map[any]any{
		"source": "raw",
		"target": "parsed",
	})
	ev := &event.Event{Data: map[string]any{"other": []byte(`{}`)}}

	_, err := f.Filter(ev)
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}
	if err.Error() != "source not found" {
		t.Fatalf("error = %q, want %q", err.Error(), "source not found")
	}
	// target 不应被写入
	if _, ok := ev.Data["parsed"]; ok {
		t.Fatal("parsed should not be created when source is missing")
	}
}

func TestJsonFilter_StringSource(t *testing.T) {
	// string 类型的 source 会被 ParseToBytes 自动转为 []byte，应正常解析
	f := newJsonFilter(map[any]any{
		"source": "raw",
		"target": "parsed",
	})
	ev := &event.Event{Data: map[string]any{"raw": `{"message":"hello"}`}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 原 string 值保持不变
	if got := ev.Data["raw"]; got != `{"message":"hello"}` {
		t.Fatalf("raw = %v, want original string", got)
	}
	got, ok := ev.Data["parsed"].(map[string]any)
	if !ok {
		t.Fatalf("parsed should be map[string]any, got %T", ev.Data["parsed"])
	}
	if got["message"] != "hello" {
		t.Fatalf("message = %v, want %q", got["message"], "hello")
	}
}

func TestJsonFilter_SourceNotBytes_IntValue(t *testing.T) {
	// 非 []byte 的其他类型同样应返回错误
	f := newJsonFilter(map[any]any{
		"source": "raw",
		"target": "parsed",
	})
	ev := &event.Event{Data: map[string]any{"raw": 123}}

	_, err := f.Filter(ev)
	if err == nil {
		t.Fatal("expected error for non-[]byte source, got nil")
	}
	if err.Error() != "invalid source" {
		t.Fatalf("error = %q, want %q", err.Error(), "invalid source")
	}
}

func TestJsonFilter_InvalidJSON(t *testing.T) {
	// 非法 JSON 字节应返回解码错误
	f := newJsonFilter(map[any]any{
		"source": "raw",
		"target": "parsed",
	})
	ev := &event.Event{Data: map[string]any{"raw": []byte(`{not json`)}}

	_, err := f.Filter(ev)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if err.Error() != "JSON decode failed" {
		t.Fatalf("error = %q, want %q", err.Error(), "JSON decode failed")
	}
	// target 不应被写入
	if _, ok := ev.Data["parsed"]; ok {
		t.Fatal("parsed should not be created when decode fails")
	}
}

func TestJsonFilter_NonObjectJSON(t *testing.T) {
	// 合法 JSON 但非对象（如数组）无法解码进 map，应返回解码错误
	f := newJsonFilter(map[any]any{
		"source": "raw",
		"target": "parsed",
	})
	ev := &event.Event{Data: map[string]any{"raw": []byte(`[1,2,3]`)}}

	_, err := f.Filter(ev)
	if err == nil {
		t.Fatal("expected error for non-object JSON, got nil")
	}
	if err.Error() != "JSON decode failed" {
		t.Fatalf("error = %q, want %q", err.Error(), "JSON decode failed")
	}
}

func TestJsonFilter_EmptyTarget_EqualsMergedMap(t *testing.T) {
	// 验证 maps.Copy 后 event.Data 的整体结构
	f := newJsonFilter(map[any]any{
		"source": "raw",
	})
	ev := &event.Event{Data: map[string]any{
		"raw": []byte(`{"a":1,"b":2}`),
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]any{
		"raw": []byte(`{"a":1,"b":2}`),
		"a":   float64(1),
		"b":   float64(2),
	}
	if !reflect.DeepEqual(ev.Data, want) {
		t.Fatalf("event.Data = %v, want %v", ev.Data, want)
	}
}

func TestNewJsonFilter_PanicsOnMissingSource(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when source missing, got none")
		}
	}()
	newJsonFilter(map[any]any{
		"target": "parsed",
	})
}

func TestNewJsonFilter_PanicsOnWrongSourceType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when source is wrong type, got none")
		}
	}()
	// int 不满足 string 断言
	newJsonFilter(map[any]any{
		"source": 123,
	})
}

func TestNewJsonFilter_PanicsOnWrongTargetType(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when target is wrong type, got none")
		}
	}()
	newJsonFilter(map[any]any{
		"source": "raw",
		"target": 123,
	})
}

func TestNewJsonFilter_DefaultTargetEmpty(t *testing.T) {
	// 仅提供 source 时，target 默认为空，Filter 走合并分支
	f := newJsonFilter(map[any]any{
		"source": "raw",
	})
	ev := &event.Event{Data: map[string]any{
		"raw": []byte(`{"message":"hello"}`),
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 合并分支：message 直接写入 event.Data
	if got := ev.Data["message"]; got != "hello" {
		t.Fatalf("message = %v, want %q", got, "hello")
	}
}

func TestBuildFilter_Json(t *testing.T) {
	// 通过注册名 "json" 构建 filter
	conf := map[any]any{
		"json": map[any]any{
			"source": "raw",
			"target": "parsed",
		},
	}
	f := BuildFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from BuildFilter")
	}
	ev := &event.Event{Data: map[string]any{
		"raw": []byte(`{"message":"BuildFilter TEST"}`),
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := ev.Data["parsed"].(map[string]any)
	if !ok {
		t.Fatalf("parsed should be map[string]any, got %T", ev.Data["parsed"])
	}
	if got["message"] != "BuildFilter TEST" {
		t.Fatalf("message = %v, want %q", got["message"], "BuildFilter TEST")
	}
}

func TestBuildFilter_JsonUnknownType(t *testing.T) {
	// 未注册的 filter 类型应返回 nil
	f := BuildFilter(map[any]any{
		"not_a_real_filter": map[any]any{},
	})
	if f != nil {
		t.Fatalf("expected nil filter for unknown type, got %T", f)
	}
}
