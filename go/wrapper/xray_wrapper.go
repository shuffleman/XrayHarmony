package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
)

// XrayInstance 表示一个 Xray 实例
type XrayInstance struct {
	server   *core.Instance
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.RWMutex
	running  bool
	config   *conf.Config
}

// Config 配置结构
type Config struct {
	InboundConfig  string `json:"inbound"`
	OutboundConfig string `json:"outbound"`
	LogLevel       string `json:"log_level"`
}

// NewXrayInstance 创建新的 Xray 实例
func NewXrayInstance() *XrayInstance {
	return &XrayInstance{
		running: false,
	}
}

// LoadConfig 从 JSON 字符串加载配置
func (x *XrayInstance) LoadConfig(configJSON string) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.running {
		return fmt.Errorf("cannot load config while instance is running")
	}

	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// 这里需要将配置转换为 xray-core 的配置格式
	// 实际实现中需要更详细的配置处理
	x.config = &conf.Config{}

	return nil
}

// LoadConfigFromFile 从文件加载配置
func (x *XrayInstance) LoadConfigFromFile(filePath string) error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.running {
		return fmt.Errorf("cannot load config while instance is running")
	}

	// 读取并解析配置文件
	// 实际实现中需要文件读取和解析逻辑
	x.config = &conf.Config{}

	return nil
}

// Start 启动 Xray 实例
func (x *XrayInstance) Start() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if x.running {
		return fmt.Errorf("instance is already running")
	}

	if x.config == nil {
		return fmt.Errorf("config not loaded")
	}

	x.ctx, x.cancel = context.WithCancel(context.Background())

	// 创建 Xray core 实例
	config, err := x.config.Build()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}

	server, err := core.New(config)
	if err != nil {
		return fmt.Errorf("failed to create core instance: %w", err)
	}

	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	x.server = server
	x.running = true

	return nil
}

// Stop 停止 Xray 实例
func (x *XrayInstance) Stop() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	if !x.running {
		return fmt.Errorf("instance is not running")
	}

	if x.cancel != nil {
		x.cancel()
	}

	if x.server != nil {
		if err := x.server.Close(); err != nil {
			return fmt.Errorf("failed to stop server: %w", err)
		}
	}

	x.running = false
	x.server = nil

	return nil
}

// IsRunning 检查实例是否正在运行
func (x *XrayInstance) IsRunning() bool {
	x.mu.RLock()
	defer x.mu.RUnlock()
	return x.running
}

// GetStats 获取统计信息
func (x *XrayInstance) GetStats() (string, error) {
	x.mu.RLock()
	defer x.mu.RUnlock()

	if !x.running {
		return "", fmt.Errorf("instance is not running")
	}

	// 获取统计信息
	stats := map[string]interface{}{
		"running": x.running,
		"status":  "ok",
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return "", fmt.Errorf("failed to marshal stats: %w", err)
	}

	return string(data), nil
}

// TestConfig 测试配置是否有效
func (x *XrayInstance) TestConfig(configJSON string) error {
	var config Config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// 验证配置的有效性
	// 实际实现中需要更详细的验证逻辑

	return nil
}
