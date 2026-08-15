package process

import (
	"strconv"
	"sync"

	"github.com/Shadowmaple/logflow/filter"
	"github.com/Shadowmaple/logflow/input"
	"github.com/Shadowmaple/logflow/internal/config"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/model"
)

type InputHandler struct {
	config  *config.Config
	input   model.Input
	outputs [][]*OutputProcessor

	stop bool
}

func buildInputHandler(config *config.Config, idx int) *InputHandler {
	if config.Input[idx] == nil {
		logger.Fatal("Input config is empty")
	}

	input := input.NewInput(config.Input[idx])
	return &InputHandler{
		config:  config,
		input:   input,
		outputs: make([][]*OutputProcessor, config.System.Worker),
	}
}

// Start 启动所有处理任务
func (h *InputHandler) start() {
	for i := 0; i < h.config.System.Worker; i++ {
		go h.startOne(i)
	}
}

// StartOne 启动一个处理任务
// 执行不退出，不断从 input 接收事件，经过 filter 处理后发送到 output
func (h *InputHandler) startOne(id int) {
	logger.Info("InputHandler startOne " + strconv.Itoa(id))

	// build filter, output processors
	filters := model.BuildFilterProcessors(h.config.Filter, filter.BuildFilter)
	outputs := BuildOutputProcessors(h.config.Output)
	if len(outputs) == 0 {
		logger.Fatal("no valid output built")
		return
	}
	h.outputs[id] = outputs

	var processor *model.ProcessNode
	for _, filter := range filters {
		processor = model.AppendProcessors(processor, filter)
	}
	for _, output := range outputs {
		processor = model.AppendProcessors(processor, output)
	}
	logger.Info("InputHandler startOne " + strconv.Itoa(id) + " processors build ok")

	// 处理事件
	for !h.stop {
		event := h.input.ReceiveOne()
		if event == nil {
			logger.Info("receive nil event, input closed")
			return
		}
		processor.Process(event)
	}

	logger.Debug("InputHandler startOne gets close signal and quickly receives all events")
	// 快速将未消费完的事件发送到下一个处理线程
	// TODO: 如果es集群挂了，链路一直卡住，那就一直无法关闭
	wg := new(sync.WaitGroup)
	for {
		event := h.input.ReceiveOne()
		if event == nil {
			break
		}
		wg.Go(func() {
			processor.Process(event)
		})
	}
	wg.Wait()
	logger.Debug("InputHandler startOne quickly receives all events ok")
}

func (h *InputHandler) close() {
	logger.Debug("InputHandler close input...")
	h.stop = true
	h.input.Close()
	h.closeOutputs()
	logger.Debug("InputHandler close input ok")
}

func (h *InputHandler) closeOutputs() {
	logger.Debug("InputHandler close all outputs...")
	wg := new(sync.WaitGroup)
	for _, outputs := range h.outputs {
		for _, output := range outputs {
			wg.Go(func() {
				output.Close()
			})
		}
	}
	wg.Wait()
	logger.Debug("InputHandler close all outputs ok")
}
