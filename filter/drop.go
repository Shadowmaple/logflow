package filter

import (
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/model"
)

type DropFilter struct{}

func init() {
	register("drop", newDropFilter)
}

func newDropFilter(config map[any]any) model.Filter {
	return &DropFilter{}
}

func (f *DropFilter) Filter(event *event.Event) (*event.Event, error) {
	return nil, nil
}
