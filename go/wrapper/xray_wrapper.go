package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/app/stats"
	coreStats "github.com/xtls/xray-core/features/stats"
)

// XrayInstance 表示一个 Xray 实例
type XrayInstance struct {
	server      *core.Instance
	statsManager coreStats.Manager
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.RWMutex
	running     bool
	config      *conf.Config
	configJSON  string
	startTime   time.Time
	assetMgr    *AssetManager
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

// NewXrayInstanceWithAssets 创建带资产管理的 Xray 实例
func NewXrayInstanceWithAssets(assetDir string) *XrayInstance {
	return &XrayInstance{
		running:  false,
		assetMgr: NewAssetManager(assetDir),
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

	// 读取配置文件
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析 JSON 配置
	return x.loadConfigFromJSON(string(data))
}

// loadConfigFromJSON 从 JSON 字符串加载配置 (内部方法)
func (x *XrayInstance) loadConfigFromJSON(configJSON string) error {
	var rawConfig map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &rawConfig); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// 转换为 xray-core 配置
	config := &conf.Config{}
	if err := json.Unmarshal([]byte(configJSON), config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	x.config = config
	x.configJSON = configJSON
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

	// 如果有资产管理器,设置资产路径
	if x.assetMgr != nil {
		// 设置环境变量以指定 geoip/geosite 路径
		geoipPath := x.assetMgr.GetAssetPath(AssetTypeGeoIP)
		geositePath := x.assetMgr.GetAssetPath(AssetTypeGeoSite)

		if _, err := os.Stat(geoipPath); err == nil {
			os.Setenv("XRAY_LOCATION_ASSET", x.assetMgr.baseDir)
		}
		if _, err := os.Stat(geositePath); err == nil {
			os.Setenv("XRAY_LOCATION_ASSET", x.assetMgr.baseDir)
		}
	}

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

	// 尝试获取统计管理器
	if statsFeature := server.GetFeature(coreStats.ManagerType()); statsFeature != nil {
		if statsMgr, ok := statsFeature.(coreStats.Manager); ok {
			x.statsManager = statsMgr
		}
	}

	x.server = server
	x.running = true
	x.startTime = time.Now()

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

	uptime := time.Since(x.startTime).Seconds()

	stats := map[string]interface{}{
		"running": x.running,
		"status":  "ok",
		"uptime":  uptime,
	}

	// 如果有统计管理器,获取流量统计
	if x.statsManager != nil {
		// 尝试获取入站和出站流量统计
		traffic := x.getTrafficStats()
		if traffic != nil {
			stats["traffic"] = traffic
		}
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return "", fmt.Errorf("failed to marshal stats: %w", err)
	}

	return string(data), nil
}

// getTrafficStats 获取流量统计 (内部方法)
func (x *XrayInstance) getTrafficStats() map[string]interface{} {
	if x.statsManager == nil {
		return nil
	}

	traffic := make(map[string]interface{})

	// 尝试获取全局上行和下行流量
	if counter := x.statsManager.GetCounter("inbound>>>traffic>>>uplink"); counter != nil {
		traffic["uplink"] = counter.Value()
	}

	if counter := x.statsManager.GetCounter("inbound>>>traffic>>>downlink"); counter != nil {
		traffic["downlink"] = counter.Value()
	}

	if len(traffic) == 0 {
		return nil
	}

	return traffic
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
