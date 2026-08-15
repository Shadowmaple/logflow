package input

import (
	"math/rand"
	"strconv"

	"github.com/Shadowmaple/logflow/codec"
	"github.com/Shadowmaple/logflow/internal/event"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"

	"go.uber.org/zap"
)

type RandomInput struct {
	decoder codec.Decoder

	from        int
	to          int
	maxMessages int
	count       int
}

func init() {
	Register("random", newRandomInput)
}

func newRandomInput(config map[any]any) model.Input {
	codecType := "plain"

	p := &RandomInput{
		decoder:     codec.NewDecoder(codecType),
		count:       0,
		maxMessages: -1,
	}

	if v, ok := config["from"]; ok {
		p.from = v.(int)
	} else {
		logger.Fatal("from must be configured in Random Input")
	}

	if v, ok := config["to"]; ok {
		p.to = v.(int)
	} else {
		logger.Fatal("to must be configured in Random Input")
	}

	if v, ok := config["max_messages"]; ok {
		p.maxMessages = v.(int)
	}

	return p
}

func (p *RandomInput) ReceiveOne() *event.Event {
	if p.maxMessages != -1 && p.count >= p.maxMessages {
		return nil
	}
	n := p.from + rand.Intn(1+p.to-p.from)
	p.count++
	res, err := p.decoder.Decode([]byte(strconv.Itoa(n)))
	if err != nil {
		logger.Error("random input: decode error:", zap.Error(err))
		return nil
	}
	return &event.Event{
		Data: res,
	}
}

func (p *RandomInput) Close() {}
