package filter

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Shadowmaple/logflow/internal/event"
)

// validDateConfig 返回一个最小可用的 date 配置（ISO8601 解析，写入 @timestamp）
func validDateConfig() map[string]any {
	return map[string]any{
		"source":  "logtime",
		"target":  "@timestamp",
		"formats": []string{"ISO8601"},
	}
}

// assertTimeEqual 断言 got 为 time.Time 且与 want 表示同一时刻
func assertTimeEqual(t *testing.T, got any, want time.Time) {
	t.Helper()
	tt, ok := got.(time.Time)
	if !ok {
		t.Fatalf("expected time.Time, got %T (%v)", got, got)
	}
	if !tt.Equal(want) {
		t.Fatalf("time = %v, want %v", tt, want)
	}
}

// --- 各解析器正常流程 ---

func TestDateFilter_ISO8601(t *testing.T) {
	f := newDateFilter(validDateConfig())
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45Z"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_RFC3339(t *testing.T) {
	cfg := validDateConfig()
	cfg["formats"] = []string{"RFC3339"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45Z"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_UNIX_String(t *testing.T) {
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": fmt.Sprintf("%d", testTime.Unix())}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_UNIX_Int(t *testing.T) {
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": int(testTime.Unix())}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_UNIX_Int64(t *testing.T) {
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": int64(testTime.Unix())}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_UNIX_FloatString(t *testing.T) {
	// 带小数的字符串：Atoi 失败后走 ParseFloat 分支，整数部分为秒、小数部分为纳秒
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX"}
	f := newDateFilter(cfg)
	src := fmt.Sprintf("%d.5", testTime.Unix())
	ev := &event.Event{Data: map[string]any{"logtime": src}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 0.5 秒 = 500000000 纳秒
	assertTimeEqual(t, ev.Data["@timestamp"], testTime.Add(500*time.Millisecond))
}

func TestDateFilter_UNIX_JsonNumber(t *testing.T) {
	// json.Number 类型（UseNumber 解码产生）走 json.Number 分支
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": json.Number(fmt.Sprintf("%d", testTime.Unix()))}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_UNIX_MS_String(t *testing.T) {
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX_MS"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": fmt.Sprintf("%d", testTime.UnixMilli())}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_UNIX_MS_Int(t *testing.T) {
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX_MS"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": int(testTime.UnixMilli())}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_CustomFormat(t *testing.T) {
	// Java 风格格式令牌被转换为 Go 格式后解析
	cfg := validDateConfig()
	cfg["formats"] = []string{"yyyy-MM-dd HH:mm:ss"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15 10:30:45"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 无时区信息时 time.Parse 返回 UTC
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

// --- 多格式回退 ---

func TestDateFilter_MultipleFormats_Fallback(t *testing.T) {
	// 第一个格式解析失败时，应尝试下一个格式
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX", "ISO8601"}
	f := newDateFilter(cfg)
	// ISO8601 字符串无法被 UNIX 解析（Atoi/ParseFloat 均失败），应回退到 ISO8601
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45Z"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_MultipleFormats_FirstSucceeds(t *testing.T) {
	// 第一个格式即可解析时，不应尝试后续格式
	cfg := validDateConfig()
	cfg["formats"] = []string{"ISO8601", "UNIX"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45Z"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

// --- target 与 overwrite ---

func TestDateFilter_DefaultTarget(t *testing.T) {
	// 未配置 target 时默认为 @timestamp
	cfg := validDateConfig()
	delete(cfg, "target")
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45Z"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_CustomTarget(t *testing.T) {
	cfg := validDateConfig()
	cfg["target"] = "mytime"
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45Z"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["mytime"], testTime)
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should not be set when target is custom")
	}
}

func TestDateFilter_OverwriteTrue_Default(t *testing.T) {
	// 默认 overwrite=true，已存在的 target 会被覆盖
	f := newDateFilter(validDateConfig())
	ev := &event.Event{Data: map[string]any{
		"logtime":    "2024-01-15T10:30:45Z",
		"@timestamp": time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_OverwriteFalse_TargetExists(t *testing.T) {
	// overwrite=false 且 target 已存在时，不覆盖
	cfg := validDateConfig()
	cfg["overwrite"] = false
	f := newDateFilter(cfg)
	old := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	ev := &event.Event{Data: map[string]any{
		"logtime":    "2024-01-15T10:30:45Z",
		"@timestamp": old,
	}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], old)
}

func TestDateFilter_OverwriteFalse_TargetAbsent(t *testing.T) {
	// overwrite=false 但 target 不存在时，正常写入
	cfg := validDateConfig()
	cfg["overwrite"] = false
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45Z"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

// --- location ---

func TestDateFilter_FormatParser_Location(t *testing.T) {
	// FormatParser 配置 location 时，按该时区解析再转 UTC
	// 输入 18:30:45 Asia/Shanghai(+08:00) -> 10:30:45 UTC == testTime
	cfg := validDateConfig()
	cfg["formats"] = []string{"yyyy-MM-dd HH:mm:ss"}
	cfg["location"] = "Asia/Shanghai"
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15 18:30:45"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}

func TestDateFilter_ISO8601_Location(t *testing.T) {
	// ISO8601 无时区信息时，使用配置的 location 解释
	// "2024-01-15T10:30:45" 解释为 Shanghai 时间 -> UTC 02:30:45
	cfg := validDateConfig()
	cfg["location"] = "Asia/Shanghai"
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}
	want := time.Date(2024, 1, 15, 10, 30, 45, 0, shanghai)
	assertTimeEqual(t, ev.Data["@timestamp"], want)
}

// --- add_year ---

func TestDateFilter_AddYear(t *testing.T) {
	// add_year=true 时，将当前年份前缀拼接到源值后再解析
	// 源值 "0115 10:30:45" + 年份 -> "20260115 10:30:45"，格式 yyyyMMdd HH:mm:ss
	cfg := validDateConfig()
	cfg["formats"] = []string{"yyyyMMdd HH:mm:ss"}
	cfg["add_year"] = true
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "0115 10:30:45"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(time.Now().Year(), 1, 15, 10, 30, 45, 0, time.UTC)
	assertTimeEqual(t, ev.Data["@timestamp"], want)
}

// --- 源字段缺失 / 异常 ---

func TestDateFilter_SourceMissing(t *testing.T) {
	// 源字段不存在时，FieldGetter 返回错误，Filter 不写入且不报错
	f := newDateFilter(validDateConfig())
	ev := &event.Event{Data: map[string]any{"other": "value"}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatalf("DateFilter should return an error")
	}
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should not be set when source is missing")
	}
}

func TestDateFilter_SourceNil(t *testing.T) {
	// 源字段存在但值为 nil 时，不写入且返回错误
	f := newDateFilter(validDateConfig())
	ev := &event.Event{Data: map[string]any{"logtime": nil}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatalf("DateFilter should return an error")
	}
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should not be set when source value is nil")
	}
}

func TestDateFilter_AllFormatsFail_NoWriteError(t *testing.T) {
	// 所有格式均无法解析时，不写入 target 且返回错误
	cfg := validDateConfig()
	cfg["formats"] = []string{"ISO8601"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": "not-a-date"}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatalf("DateFilter should return an error")
	}
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should not be set when all formats fail")
	}
}

func TestDateFilter_SourceUnsupportedType(t *testing.T) {
	// 源值类型无解析器可处理（如 bool）时，不写入且返回错误
	cfg := validDateConfig()
	cfg["formats"] = []string{"UNIX"}
	f := newDateFilter(cfg)
	ev := &event.Event{Data: map[string]any{"logtime": true}}

	if _, err := f.Filter(ev); err == nil {
		t.Fatalf("DateFilter should return an error")
	}
	if _, exists := ev.Data["@timestamp"]; exists {
		t.Fatal("@timestamp should not be set for unsupported source type")
	}
}

// --- 构造函数 panic ---

func TestNewDateFilter_PanicsOnMissingSource(t *testing.T) {
	cfg := validDateConfig()
	delete(cfg, "source")
	assertPanics(t, func() { newDateFilter(cfg) })
}

func TestNewDateFilter_PanicsOnMissingFormats(t *testing.T) {
	cfg := validDateConfig()
	delete(cfg, "formats")
	assertPanics(t, func() { newDateFilter(cfg) })
}

func TestNewDateFilter_PanicsOnEmptyFormats(t *testing.T) {
	cfg := validDateConfig()
	cfg["formats"] = []string{}
	assertPanics(t, func() { newDateFilter(cfg) })
}

func TestNewDateFilter_PanicsOnInvalidLocation(t *testing.T) {
	cfg := validDateConfig()
	cfg["location"] = "Invalid/Nonexistent/Zone"
	assertPanics(t, func() { newDateFilter(cfg) })
}

// --- 集成 ---

func TestBuildFilter_Date(t *testing.T) {
	// 通过注册名 "Date" 构建 filter
	conf := map[string]any{
		"date": map[string]any{
			"source":  "logtime",
			"target":  "@timestamp",
			"formats": []string{"ISO8601"},
		},
	}
	f := BuildFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from BuildFilter")
	}
	ev := &event.Event{Data: map[string]any{"logtime": "2024-01-15T10:30:45Z"}}

	if _, err := f.Filter(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTimeEqual(t, ev.Data["@timestamp"], testTime)
}
