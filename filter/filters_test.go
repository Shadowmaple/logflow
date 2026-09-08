package filter

import (
	"testing"

	"github.com/Shadowmaple/logflow/internal/event"
)

func TestFilters(t *testing.T) {
	conf := map[any]any{
		"modules": []any{
			map[any]any{
				"strip": map[any]any{
					"fields": []any{"level"},
				},
			},
			map[any]any{
				"uppercase": map[any]any{
					"fields": []any{"level"},
				},
			},
		},
	}
	filter := newFiltersFilter(conf)
	if filter == nil {
		t.Fatalf("expected non-nil filter, got nil")
	}
	event := &event.Event{
		Data: map[string]any{
			"level": "   info    ",
		},
	}
	newEvent, err := filter.Filter(event)
	if err != nil {
		t.Fatalf("filter: %v", err)
	}
	if newEvent.Data["level"] != "INFO" {
		t.Fatalf("expected level to be uppercase, got %v", newEvent.Data["level"])
	}
}
