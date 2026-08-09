package filter

import (
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/model"
)

type DropFilter struct{}

func init() {
	register("drop", newDropFilter)
}

func newDropFilter(config map[string]any) model.Filter {
	return &DropFilter{}
}

func (f *DropFilter) Filter(event *event.Event) error {
	event = nil
	return nil
}
