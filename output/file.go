package output

import "github.com/Shadowmaple/logflow/internal/model"

type FileOutput struct{}

func (f *FileOutput) Handle(event *model.Event) error {
	return nil
}
