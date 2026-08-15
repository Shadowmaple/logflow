package filter

import (
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

func TestDrop(t *testing.T) {
	conf := map[any]any{
		"if": []string{`EQ(level, "DEBUG")`},
	}
	f := newDropFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newDropFilter")
	}
	event := &event.Event{
		Data: map[string]any{
			"level": "DEBUG",
		},
	}
	if newEvent, err := f.Filter(event); err != nil {
		t.Errorf("drop filter failed: %v", err)
	} else if newEvent != nil {
		t.Errorf("drop filter failed: %v", newEvent)
	}
}
