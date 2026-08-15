package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type LoggerConfig struct {
	Level    string `yaml:"level"`     // 日志级别: debug, info, warn, error, fatal
	Format   string `yaml:"format"`    // 输出格式: json, console
	Output   string `yaml:"output"`    // 输出位置: stdout, stderr, file
	FilePath string `yaml:"file_path"` // 输出到文件时的文件路径
}

type SystemConfig struct {
	Worker int           `yaml:"worker"` // 并发处理数，即每个input开启n个任务协程并发处理，每个input都有单独的filter和output队列
	Logger *LoggerConfig `yaml:"logger"`
}

type Config struct {
	System *SystemConfig `yaml:"system"`

	Input  []map[any]any `yaml:"input"`
	Filter []map[any]any `yaml:"filter"`
	Output []map[any]any `yaml:"output"`
}

func NewConfig() *Config {
	return &Config{
		System: &SystemConfig{
			Worker: 1,
		},
	}
}

// parse config file in .yaml or .yml format
func LoadConfig(path string) (*Config, error) {
	log.Printf("Logflow loads config from: %s", path)

	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file %s not exist", path)
	}

	// 检查文件后缀是否为.yaml或.yml
	if ext := filepath.Ext(path); ext != ".yaml" && ext != ".yml" {
		return nil, fmt.Errorf("config must be yaml file")
	}

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read conf failed: %w", err)
	}

	var cfg = NewConfig()
	if err = yaml.Unmarshal(yamlFile, cfg); err != nil {
		return nil, fmt.Errorf("unmarshal conf failed: %w", err)
	}
	// 设置默认值
	cfg.SetDefault()
	return cfg, nil
}

func (c *Config) SetDefault() {
	c.System.Worker = max(c.System.Worker, 1)

	if c.System.Logger == nil {
		c.System.Logger = &LoggerConfig{}
	}
	if c.System.Logger.Level == "" {
		c.System.Logger.Level = "info"
	}
	if c.System.Logger.Format == "" {
		c.System.Logger.Format = "json"
	}
	if c.System.Logger.Output == "" {
		c.System.Logger.Output = "stdout"
	}
}
