package model

import (
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"

	"go.uber.org/zap"
)

type Processor interface {
	Process(*event.Event) *event.Event
}

type ProcessNode struct {
	P    Processor
	Next *ProcessNode
}

// func NewProcessList(processors ...Processor) *ProcessNode {
// 	head := &ProcessNode{}
// 	if len(processors) > 0 {
// 		head.Append(processors...)
// 	}
// 	return head.Next
// }

func AppendProcessors(head *ProcessNode, processors ...Processor) *ProcessNode {
	newHead := &ProcessNode{Next: head}
	cur := newHead
	for ; cur.Next != nil; cur = cur.Next {
	}

	for _, p := range processors {
		cur.Next = &ProcessNode{P: p}
		cur = cur.Next
	}
	return newHead.Next
}

func (pn *ProcessNode) Process(event *event.Event) *event.Event {
	if event = pn.P.Process(event); event == nil {
		logger.Error("process event failed", zap.String("event", event.String()))
	}
	if event != nil && pn.Next != nil {
		return pn.Next.Process(event)
	}
	return event
}

// func (pn *ProcessNode) Append(processors ...Processor) {
// 	if pn == nil {
// 		logger.Warn("append processor to nil node")
// 		return
// 	}
// 	cur := pn
// 	for ; cur.Next != nil; cur = cur.Next {
// 	}

// 	for _, p := range processors {
// 		cur.Next = &ProcessNode{P: p}
// 		cur = cur.Next
// 	}
// }
