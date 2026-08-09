package model

import "github.com/Shadowmaple/logflow/internal/event"

type Input interface {
	// Receive() <-chan *event.Event
	ReceiveOne() *event.Event
	Close()
}
