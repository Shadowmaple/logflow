package output

import (
	"fmt"

	"github.com/Shadowmaple/logflow/codec"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/utils"
	"github.com/Shadowmaple/logflow/model"

	"go.uber.org/zap"
)

type ConsoleOutput struct {
	encoder codec.Encoder
}

func init() {
	Register("console", newConsoleOutput)
}

func newConsoleOutput(config map[any]any) model.Output {
	codecType := "plain"
	if codecCfg, ok := config["codec"]; ok {
		if codecType, ok = utils.ParseToStr(codecCfg); !ok {
			panic("console output: config invalid codec")
		}
	}

	return &ConsoleOutput{
		encoder: codec.NewEncoder(codecType),
	}
}

func (p *ConsoleOutput) Handle(event *event.Event) error {
	buf, err := p.encoder.Encode(event.Data)
	if err != nil {
		logger.Error("console output: encode error:", zap.Error(err))
		return err
	}
	fmt.Println(string(buf))
	return nil
}

func (p *ConsoleOutput) Close() {}
