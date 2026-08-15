package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shadowmaple/logflow/internal/config"
	"github.com/Shadowmaple/logflow/internal/logger"
	"github.com/Shadowmaple/logflow/process"
)

var options struct {
	config string
}

func main() {
	// flags
	flag.StringVar(&options.config, "c", options.config, "path to configuration file or directory")

	flag.Parse()

	// 加载配置文件
	conf, err := config.LoadConfig(options.config)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	// init logger
	logger.Init(conf.Logger)
	defer logger.Sync()

	// 用于接收退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建Input并启动
	inputGroup := process.BuildInputGroup(conf)
	go inputGroup.Start()

	<-sigChan
	logger.Info("Logflow received signal to exit")
	inputGroup.Close()
	logger.Info("Logflow exit ok")
}
