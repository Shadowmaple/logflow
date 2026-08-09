package model

import "github.com/Shadowmaple/logflow/internal/event"

type Output interface {
	Handle(event *event.Event) error
	Close()
}
