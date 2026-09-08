package filter

import (
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

func TestStrip(t *testing.T) {
	conf := map[any]any{
		"fields": []any{"level"},
	}
	f := newStripFilter(conf)
	if f == nil {
		t.Fatal("expected non-nil filter from newStripFilter")
	}

	events := []event.Event{
		{
			Data: map[string]any{
				"level": "  DEBUG  ",
			},
		},
		{
			Data: map[string]any{
				"level": "  INFO  ",
			},
		},
		{
			Data: map[string]any{
				"level": "  WARN  \n",
			},
		},
	}
	expected := []string{
		"DEBUG",
		"INFO",
		"WARN",
	}
	for i, event := range events {
		if _, err := f.Filter(&event); err != nil {
			t.Errorf("strip filter failed: %v", err)
		}
		if event.Data["level"] != expected[i] {
			t.Errorf("strip filter failed: %v", event.Data["level"])
		}
	}
}
