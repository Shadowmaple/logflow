package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Worker int              `yaml:"worker"` // 处理线程数
	Logger map[string]any   `yaml:"logger"`
	Input  []map[string]any `yaml:"input"`
	Filter []map[string]any `yaml:"filter"`
	Output []map[string]any `yaml:"output"`
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

	var cfg Config
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read conf failed: %w", err)
	}

	if err = yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal conf failed: %w", err)
	}
	return &cfg, nil
}
