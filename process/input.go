package process

import (
	"sync"

	"github.com/Shadowmaple/logflow/input"
	"github.com/Shadowmaple/logflow/internal/config"
	"github.com/Shadowmaple/logflow/internal/logger"
)

type InputHandler struct {
	config  *config.Config
	input   input.Input
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
		outputs: make([][]*OutputProcessor, config.Worker),
	}
}

// Start 启动所有处理任务
func (h *InputHandler) start() {
	for i := 0; i < h.config.Worker; i++ {
		go h.startOne(i)
	}
}

// StartOne 启动一个处理任务
// 执行不退出，不断从 input 接收事件，经过 filter 处理后发送到 output
func (h *InputHandler) startOne(id int) {
	// build filter, output processors
	filters := BuildFilterProcessors(h.config.Filter)
	outputs := BuildOutputProcessors(h.config.Output)
	if len(outputs) == 0 {
		logger.Fatal("no valid output built")
		return
	}
	h.outputs[id] = outputs

	processor := NewProcessList()
	for _, filter := range filters {
		processor.Append(filter)
	}
	for _, output := range outputs {
		processor.Append(output)
	}

	// 处理事件
	for !h.stop {
		event := h.input.ReceiveOne()
		if event == nil {
			logger.Info("receive nil event, input closed")
			return
		}
		processor.Process(event)
	}

	// 快速将未消费完的事件发送到下一个处理线程
	// TODO: 如果es集群挂了，链路一直卡住，那就一直无法关闭
	wg := new(sync.WaitGroup)
	for {
		event := h.input.ReceiveOne()
		if event == nil {
			break
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			processor.Process(event)
		}()
	}
	wg.Wait()
}

func (h *InputHandler) close() {
	h.stop = true
	h.input.Close()
	h.closeOutputs()
}

func (h *InputHandler) closeOutputs() {
	wg := new(sync.WaitGroup)
	for _, outputs := range h.outputs {
		for _, output := range outputs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				output.Close()
			}()
		}
	}
	wg.Wait()
	logger.Info("InputHandler close all outputs ok")
}
