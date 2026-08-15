package input

import (
	"bufio"
	"os"
	"sync"
	"time"

	"github.com/Shadowmaple/logflow/codec"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/internal/utils"
	"github.com/Shadowmaple/logflow/model"

	"go.uber.org/zap"
)

type ConsoleInput struct {
	decoder codec.Decoder

	scanner  *bufio.Scanner
	scanLock sync.Mutex

	// stop bool
}

func init() {
	Register("console", newConsoleInput)
}

func newConsoleInput(config map[any]any) model.Input {
	codecType := "plain"
	if codecCfg, ok := config["codec"]; ok {
		if codecType, ok = utils.ParseToStr(codecCfg); !ok {
			panic("console input: config invalid codec")
		}
	}

	return &ConsoleInput{
		decoder: codec.NewDecoder(codecType),
		scanner: bufio.NewScanner(os.Stdin),
	}
}

func (p *ConsoleInput) ReceiveOne() *event.Event {
	p.scanLock.Lock()
	defer p.scanLock.Unlock()

	if p.scanner.Scan() {
		t := p.scanner.Bytes()
		msg := make([]byte, len(t))
		copy(msg, t)
		res, err := p.decoder.Decode(msg)
		if err != nil {
			logger.Error("console input: decode error:", zap.Error(err))
			return nil
		}
		return &event.Event{
			Data: res,
		}
	}
	if err := p.scanner.Err(); err != nil {
		logger.Error("console input: scan error:", zap.Error(err))
		return nil
	} else {
		// EOF here. when stdin is closed by C-D, cpu will raise up to 100% if not sleep
		time.Sleep(time.Millisecond * 1000)
	}
	return nil
}

func (p *ConsoleInput) Close() {}
