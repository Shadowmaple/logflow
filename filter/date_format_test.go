package filter

import (
	"testing"
	"time"

	"github.com/Shadowmaple/logflow/internal/event"
)

// validDateFormatConfig 返回一个最小可用的 dateFormat 配置
func validDateFormatConfig() map[any]any {
	return map[any]any{
		"source": "logtime",
		"target": "@timestamp",
		"format": "yyyy-MM-dd HH:mm:ss",
	}
}

// testTime 是一个固定的测试时间点，位于 UTC
var testTime = time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)

// assertPanics 断言 fn 执行时会 panic
func assertPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic, got none")
		}
	}()
	fn()
}

func TestDateFormatFilter_BasicFormat(t *testing.T) {
	f := newDateFormatFilter(validDateFormatConfig())
	ev := &event.Event{Data: map[string]any{"logtime": testTime}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["@timestamp"]; got != "2024-01-15 10:30:45" {
		t.Fatalf("@timestamp = %v, want %q", got, "2024-01-15 10:30:45")
	}
}

func TestDateFormatFilter_JavaFormatConversion(t *testing.T) {
	// 验证 Java 风格格式令牌被正确转换为 Go 格式
	cfg := validDateFormatConfig()
	cfg["format"] = "yyyy/MM/dd"
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": testTime}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["@timestamp"]; got != "2024/01/15" {
		t.Fatalf("@timestamp = %v, want %q", got, "2024/01/15")
	}
}

func TestDateFormatFilter_SourcePointerToTime(t *testing.T) {
	// *time.Time 类型的源值应当能被正确解析
	t1 := testTime
	f := newDateFormatFilter(validDateFormatConfig())
	ev := &event.Event{Data: map[string]any{"logtime": &t1}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["@timestamp"]; got != "2024-01-15 10:30:45" {
		t.Fatalf("@timestamp = %v, want %q", got, "2024-01-15 10:30:45")
	}
}

func TestDateFormatFilter_Location(t *testing.T) {
	// 配置 location 后，时间应先转换到该时区再格式化
	// testTime 为 UTC 10:30:45，Asia/Shanghai 为 +08:00，应为 18:30:45
	cfg := validDateFormatConfig()
	cfg["location"] = "Asia/Shanghai"
	cfg["format"] = "HH:mm:ss"
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": testTime}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["@timestamp"]; got != "18:30:45" {
		t.Fatalf("@timestamp = %v, want %q", got, "18:30:45")
	}
}

func TestDateFormatFilter_SourceNotExists_SetIfNil(t *testing.T) {
	// 源字段不存在且配置了 set_if_nil 时，写入默认值，但仍返回错误
	cfg := validDateFormatConfig()
	cfg["set_if_nil"] = "1970-01-01 00:00:00"
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := ev.Data["@timestamp"]; got != "1970-01-01 00:00:00" {
		t.Fatalf("@timestamp = %v, want %q", got, "1970-01-01 00:00:00")
	}
}

func TestDateFormatFilter_SourceNotExists_NoSetIfNil(t *testing.T) {
	// 源字段不存在且未配置 set_if_nil 时，不写入 target，返回错误
	f := newDateFormatFilter(validDateFormatConfig())
	ev := &event.Event{Data: map[string]any{}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should not be set when source is missing and set_if_nil is empty")
	}
}

func TestDateFormatFilter_SourceNotTime_SetIfFail(t *testing.T) {
	// 源字段存在但不是 time.Time，配置 set_if_fail 时写入回退值，但仍返回错误
	cfg := validDateFormatConfig()
	cfg["set_if_fail"] = "fallback"
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "not-a-time"}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := ev.Data["@timestamp"]; got != "fallback" {
		t.Fatalf("@timestamp = %v, want %q", got, "fallback")
	}
}

func TestDateFormatFilter_SourceNotTime_RemoveIfFail(t *testing.T) {
	// 源字段不是 time.Time 且 remove_if_fail=true 时，删除已存在的 target，但仍返回错误
	cfg := validDateFormatConfig()
	cfg["remove_if_fail"] = true
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "not-a-time", "@timestamp": "old"}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should be removed when remove_if_fail is true")
	}
}

func TestDateFormatFilter_SourceNotTime_RemoveIfFail_TargetAbsent(t *testing.T) {
	// remove_if_fail=true 但 target 本就不存在时，不应 panic，返回错误
	cfg := validDateFormatConfig()
	cfg["remove_if_fail"] = true
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "not-a-time"}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should remain absent")
	}
}

func TestDateFormatFilter_SourceNotTime_NoFallback(t *testing.T) {
	// 源字段不是 time.Time，且未配置 set_if_fail / remove_if_fail 时，什么都不做，返回错误
	f := newDateFormatFilter(validDateFormatConfig())
	ev := &event.Event{Data: map[string]any{"logtime": 12345}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should not be set when no fallback is configured")
	}
}

func TestDateFormatFilter_SourceNilValue_SetIfFail(t *testing.T) {
	// 源字段值为 nil 时，视为提取失败，触发 set_if_fail，返回错误
	cfg := validDateFormatConfig()
	cfg["set_if_fail"] = "fallback"
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": nil}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := ev.Data["@timestamp"]; got != "fallback" {
		t.Fatalf("@timestamp = %v, want %q", got, "fallback")
	}
}

func TestDateFormatFilter_SourceNilPointer_SetIfFail(t *testing.T) {
	// 源字段为 nil 的 *time.Time 时，提取失败，触发 set_if_fail，返回错误
	var t1 *time.Time
	cfg := validDateFormatConfig()
	cfg["set_if_fail"] = "fallback"
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": t1}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := ev.Data["@timestamp"]; got != "fallback" {
		t.Fatalf("@timestamp = %v, want %q", got, "fallback")
	}
}

func TestDateFormatFilter_OverwriteTrue_Default(t *testing.T) {
	// 默认 overwrite=true，已存在的 target 会被覆盖
	f := newDateFormatFilter(validDateFormatConfig())
	ev := &event.Event{Data: map[string]any{"logtime": testTime, "@timestamp": "old"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["@timestamp"]; got != "2024-01-15 10:30:45" {
		t.Fatalf("@timestamp = %v, want %q", got, "2024-01-15 10:30:45")
	}
}

func TestDateFormatFilter_OverwriteFalse_TargetExists(t *testing.T) {
	// overwrite=false 且 target 已存在时，不覆盖
	cfg := validDateFormatConfig()
	cfg["overwrite"] = false
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": testTime, "@timestamp": "existing"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["@timestamp"]; got != "existing" {
		t.Fatalf("@timestamp = %v, want %q (should not be overwritten)", got, "existing")
	}
}

func TestDateFormatFilter_OverwriteFalse_TargetAbsent(t *testing.T) {
	// overwrite=false 但 target 不存在时，正常写入
	cfg := validDateFormatConfig()
	cfg["overwrite"] = false
	f := newDateFormatFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": testTime}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["@timestamp"]; got != "2024-01-15 10:30:45" {
		t.Fatalf("@timestamp = %v, want %q", got, "2024-01-15 10:30:45")
	}
}

func TestNewDateFormatFilter_PanicsOnMissingSource(t *testing.T) {
	cfg := validDateFormatConfig()
	delete(cfg, "source")
	assertPanics(t, func() { newDateFormatFilter(cfg) })
}

func TestNewDateFormatFilter_PanicsOnMissingTarget(t *testing.T) {
	cfg := validDateFormatConfig()
	delete(cfg, "target")
	assertPanics(t, func() { newDateFormatFilter(cfg) })
}

func TestNewDateFormatFilter_PanicsOnMissingFormat(t *testing.T) {
	cfg := validDateFormatConfig()
	delete(cfg, "format")
	assertPanics(t, func() { newDateFormatFilter(cfg) })
}

func TestNewDateFormatFilter_PanicsOnInvalidLocation(t *testing.T) {
	cfg := validDateFormatConfig()
	cfg["location"] = "Invalid/Nonexistent/Zone"
	assertPanics(t, func() { newDateFormatFilter(cfg) })
}

func TestBuildFilter_DateFormat(t *testing.T) {
	// 通过注册名 "dateFormat" 构建 filter
	conf := map[any]any{
		"dateFormat": map[any]any{
			"source": "logtime",
			"target": "@timestamp",
			"format": "yyyy-MM-dd",
		},
	}
	f := BuildFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from BuildFilter")
	}
	ev := &event.Event{Data: map[string]any{"logtime": testTime}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ev.Data["@timestamp"]; got != "2024-01-15" {
		t.Fatalf("@timestamp = %v, want %q", got, "2024-01-15")
	}
}
