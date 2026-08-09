package process

import (
	"fmt"
	"sync"

	"github.com/Shadowmaple/logflow/internal/config"
	"github.com/Shadowmaple/logflow/internal/logger"
)

// type InputGroup struct {
// 	Conf   *config.Config
// 	Inputs []*InputHandler

// 	closeCh  chan struct{} // 接收关闭信号
// 	closedCh chan struct{} // 通知关闭完成
// }

type InputGroup []*InputHandler

func BuildInputGroup(config *config.Config) InputGroup {
	if len(config.Input) == 0 {
		panic("no input config")
	}
	inputs := make([]*InputHandler, 0, len(config.Input))
	for idx := range config.Input {
		inputs = append(inputs, buildInputHandler(config, idx))
	}
	if len(inputs) == 0 {
		panic("no valid input built")
	}

	logger.Info(fmt.Sprintf("%d input build ok", len(inputs)))

	return InputGroup(inputs)
}

func (ig InputGroup) Start() {
	// wg := new(sync.WaitGroup)
	for _, input := range ig {
		// wg.Add(1)
		// go func() {
		// 	defer wg.Done()
		// 	input.start()
		// }()
		go input.start()
	}
	// wg.Wait()
}

// // Start 启动所有input
// func (ig *InputGroup) Start() {
// 	ctx, cancel := context.WithCancel(context.Background())

// 	wg := new(sync.WaitGroup)
// 	for _, input := range ig.Inputs {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			ig.startOne(ctx, input)
// 		}()
// 	}

// 	<-ig.closeCh
// 	cancel()
// 	// 等待每个input完成退出
// 	wg.Wait()
// 	close(ig.closedCh)
// 	logger.Info("InputGroup Start closed")
// }

func (ig InputGroup) Close() {
	logger.Info("InputGroup is closing")
	wg := new(sync.WaitGroup)
	// 关闭input
	for _, input := range ig {
		wg.Add(1)
		go func() {
			defer wg.Done()
			input.close()
		}()
	}
	wg.Wait()
	logger.Info("InputGroup Close ok")
}

// // startOne 启动一个input
// func (ig *InputGroup) startOne(ctx context.Context, input input.Input) {
// 	// 启动多个worker
// 	for i := 0; i < ig.Conf.Worker; i++ {
// 		ig.startOneWorker(i)
// 	}
// 	// // build filter, output processors
// 	// filters := BuildFilterProcessors(ig.Conf.Filter)
// 	// outputs := BuildOutputProcessors(ig.Conf.Output)
// 	// if len(outputs) == 0 {
// 	// 	logger.Fatal("no valid output built")
// 	// 	return
// 	// }
// 	// // close all outputs
// 	// defer ig.closeOutputs(outputs)

// 	// processor := NewProcessList()
// 	// for _, filter := range filters {
// 	// 	processor.Append(filter)
// 	// }
// 	// for _, output := range outputs {
// 	// 	processor.Append(output)
// 	// }

// 	// waitCh := make(chan struct{})
// 	// // 消费主协程
// 	// go func() {
// 	// 	defer close(waitCh)
// 	// 	for {
// 	// 		select {
// 	// 		case <-ctx.Done():
// 	// 			logger.Info("input goroutine exit by ctx")
// 	// 			return
// 	// 		case event, ok := <-input.Receive():
// 	// 			if !ok {
// 	// 				logger.Info("input channel closed")
// 	// 				return
// 	// 			}
// 	// 			processor.Process(event)
// 	// 		}
// 	// 	}
// 	// }()

// 	// // 等待input消费主协程退出
// 	// <-waitCh

// 	// wg := new(sync.WaitGroup)
// 	// // 快速消费channel中未处理的event
// 	// for event := range input.Receive() {
// 	// 	wg.Add(1)
// 	// 	go func(event event.Event) {
// 	// 		defer wg.Done()
// 	// 		processor.Process(&event)
// 	// 	}(*event)
// 	// }
// 	// wg.Wait()
// }

// // startOneWorker 启动一个input的worker
// func (ig *InputGroup) startOneWorker(id int) {
// 	// build filter, output processors
// 	filters := BuildFilterProcessors(ig.Conf.Filter)
// 	outputs := BuildOutputProcessors(ig.Conf.Output)
// 	if len(outputs) == 0 {
// 		logger.Fatal("no valid output built")
// 		return
// 	}
// 	// close all outputs
// 	defer ig.closeOutputs(outputs)

// 	processor := NewProcessList()
// 	for _, filter := range filters {
// 		processor.Append(filter)
// 	}
// 	for _, output := range outputs {
// 		processor.Append(output)
// 	}

// 	waitCh := make(chan struct{})
// 	// 消费主协程
// 	go func() {
// 		defer close(waitCh)
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				logger.Info("input goroutine exit by ctx")
// 				return
// 			case event, ok := <-input.Receive():
// 				if !ok {
// 					logger.Info("input channel closed")
// 					return
// 				}
// 				processor.Process(event)
// 			}
// 		}
// 	}()

// 	// 等待input消费主协程退出
// 	<-waitCh

// 	wg := new(sync.WaitGroup)
// 	// 快速消费channel中未处理的event
// 	for event := range input.Receive() {
// 		wg.Add(1)
// 		go func(event event.Event) {
// 			defer wg.Done()
// 			processor.Process(&event)
// 		}(*event)
// 	}
// 	wg.Wait()
// }

// func (ig InputGroup) closeOutputs(outputs []*OutputProcessor) {
// 	wg := new(sync.WaitGroup)

// 	for _, output := range outputs {
// 		wg.Add(1)
// 		go func(o *OutputProcessor) {
// 			defer wg.Done()
// 			o.Close()
// 		}(output)
// 	}

// 	wg.Wait()
// }
