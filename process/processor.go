package process

// type ProcessNode struct {
// 	p    model.Processor
// 	next *ProcessNode
// }

// func NewProcessList(processors ...model.Processor) *ProcessNode {
// 	head := &ProcessNode{}
// 	if len(processors) > 0 {
// 		head.Append(processors...)
// 	}
// 	return head.next
// }

// func (pn *ProcessNode) Process(event *event.Event) {
// 	if !pn.p.Process(event) {
// 		logger.Error("process event failed", zap.String("event", event.String()))
// 		return
// 	}
// 	if event != nil && pn.next != nil {
// 		pn.next.Process(event)
// 	}
// }

// func (pn *ProcessNode) Append(processors ...model.Processor) {
// 	cur := &ProcessNode{next: pn}
// 	for ; cur.next != nil; cur = cur.next {
// 	}

// 	for _, p := range processors {
// 		cur.next = &ProcessNode{p: p}
// 		cur = cur.next
// 	}
// }
