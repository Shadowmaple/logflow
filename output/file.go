package output

import "github.com/Shadowmaple/logflow/internal/event"

type FileOutput struct{}

func (f *FileOutput) Handle(event *event.Event) error {
	return nil
}
