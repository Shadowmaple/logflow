package field_setter

import (
	"reflect"
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

// newFieldEvent 构造一个带初始数据的测试事件；data 为 nil 时初始化为空 map
func newFieldEvent(data map[string]any) *event.Event {
	if data == nil {
		data = map[string]any{}
	}
	return &event.Event{Data: data}
}

// assertDataEqual 断言 event.Data 与 want 深度相等
func assertDataEqual(t *testing.T, ev *event.Event, want map[string]any) {
	t.Helper()
	if !reflect.DeepEqual(ev.Data, want) {
		t.Fatalf("Data = %v, want %v", ev.Data, want)
	}
}

// ===============
// NewFieldSetter
// ===============

func TestNewFieldSetter_PlainString_ReturnsOneLevel(t *testing.T) {
	// 无括号的模板应返回 OneLevelFieldSetter，字段名即模板原文
	f := NewFieldSetter("logtime", true)
	if _, ok := f.(*OneLevelFieldSetter); !ok {
		t.Fatalf("expected *OneLevelFieldSetter, got %T", f)
	}
	ev := newFieldEvent(nil)
	f.SetField(ev, "v")
	if got, ok := ev.Data["logtime"]; !ok || got != "v" {
		t.Fatalf("logtime = %v, want %q", got, "v")
	}
}

func TestNewFieldSetter_SingleBracket_ReturnsOneLevel(t *testing.T) {
	// 单层括号 "[logtime]" 应剥离括号，等价于单层字段 "logtime"
	f := NewFieldSetter("[logtime]", true)
	if _, ok := f.(*OneLevelFieldSetter); !ok {
		t.Fatalf("expected *OneLevelFieldSetter, got %T", f)
	}
	ev := newFieldEvent(nil)
	f.SetField(ev, "v")
	// 应写入 "logtime"，而非 "[logtime]"
	if got, ok := ev.Data["logtime"]; !ok || got != "v" {
		t.Fatalf("logtime = %v, want %q", got, "v")
	}
	if _, exists := ev.Data["[logtime]"]; exists {
		t.Fatal(`should not write key "[logtime]"`)
	}
}

func TestNewFieldSetter_MultipleBrackets_ReturnsMultiLevel(t *testing.T) {
	// 多层括号 "[a][b]" 应返回 MultiLevelFieldSetter
	f := NewFieldSetter("[a][b]", true)
	if _, ok := f.(*MultiLevelFieldSetter); !ok {
		t.Fatalf("expected *MultiLevelFieldSetter, got %T", f)
	}
}

func TestNewFieldSetter_OverwritePropagated(t *testing.T) {
	// overwrite 标志应透传给底层 setter：字段已存在且 overwrite=false 时不覆盖
	f := NewFieldSetter("message", false)
	ev := newFieldEvent(map[string]any{"message": "old"})
	f.SetField(ev, "new")
	if got := ev.Data["message"]; got != "old" {
		t.Fatalf("message = %v, want %q (overwrite=false should skip)", got, "old")
	}
}

func TestNewFieldSetter_OverwriteTruePropagated(t *testing.T) {
	// overwrite=true 时应覆盖已有字段
	f := NewFieldSetter("message", true)
	ev := newFieldEvent(map[string]any{"message": "old"})
	f.SetField(ev, "new")
	if got := ev.Data["message"]; got != "new" {
		t.Fatalf("message = %v, want %q (overwrite=true should overwrite)", got, "new")
	}
}
