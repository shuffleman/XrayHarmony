package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/xjasonlyu/tun2socks/v2/core"
	"github.com/xjasonlyu/tun2socks/v2/core/device"
	"github.com/xjasonlyu/tun2socks/v2/core/device/fdbased"
	"github.com/xjasonlyu/tun2socks/v2/proxy"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// VPNConfig VPN 配置结构
type VPNConfig struct {
	TunFd         int      `json:"tunFd"`         // TUN 设备文件描述符
	TunMTU        int      `json:"tunMTU"`        // TUN MTU
	SocksAddr     string   `json:"socksAddr"`     // Xray SOCKS5 地址
	DNSServers    []string `json:"dnsServers"`    // DNS 服务器列表
	FakeDNS       bool     `json:"fakeDNS"`       // 是否启用 FakeDNS
	UDP           bool     `json:"udp"`           // 是否启用 UDP
	TCPConcurrent bool     `json:"tcpConcurrent"` // 是否启用 TCP 并发
}

// VPNManager VPN 管理器
type VPNManager struct {
	mu         sync.RWMutex
	running    bool
	config     *VPNConfig
	tunDevice  device.Device
	ctx        context.Context
	cancel     context.CancelFunc
	stack      *stack.Stack
	proxy      proxy.Proxy
	xrayClient *XrayInstance
}

// NewVPNManager 创建新的 VPN 管理器
func NewVPNManager(xrayInstance *XrayInstance) *VPNManager {
	return &VPNManager{
		running:    false,
		xrayClient: xrayInstance,
	}
}

// Start 启动 VPN
func (vm *VPNManager) Start(config *VPNConfig) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.running {
		return fmt.Errorf("VPN is already running")
	}

	// 验证配置
	if err := vm.validateConfig(config); err != nil {
		return fmt.Errorf("invalid VPN config: %w", err)
	}

	vm.config = config
	vm.ctx, vm.cancel = context.WithCancel(context.Background())

	// 1. 创建 TUN 设备（使用文件描述符）
	mtu := uint32(config.TunMTU)
	if mtu == 0 {
		mtu = 1500
	}

	tunDevice, err := fdbased.Open(fmt.Sprintf("fd://%d", config.TunFd), mtu)
	if err != nil {
		return fmt.Errorf("failed to open TUN device: %w", err)
	}
	vm.tunDevice = tunDevice

	// 2. 创建代理客户端（连接到 Xray SOCKS5）
	vm.proxy, err = proxy.NewSocks5Proxy(config.SocksAddr)
	if err != nil {
		tunDevice.Close()
		return fmt.Errorf("failed to create proxy: %w", err)
	}

	// 3. 启动网络栈
	if err := vm.startStack(); err != nil {
		tunDevice.Close()
		return fmt.Errorf("failed to start network stack: %w", err)
	}

	vm.running = true
	return nil
}

// startStack 启动网络栈
func (vm *VPNManager) startStack() error {
	// 创建网络栈
	s, err := core.CreateStack(&core.Config{
		LinkEndpoint: vm.tunDevice,
		Handler:      vm.proxy,
		UDPTimeout:   60,
	})
	if err != nil {
		return fmt.Errorf("failed to create stack: %w", err)
	}

	vm.stack = s
	return nil
}

// Stop 停止 VPN
func (vm *VPNManager) Stop() error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if !vm.running {
		return fmt.Errorf("VPN is not running")
	}

	// 取消上下文
	if vm.cancel != nil {
		vm.cancel()
	}

	// 停止网络栈
	if vm.stack != nil {
		vm.stack.Close()
		vm.stack.Wait()
		vm.stack = nil
	}

	// 关闭 TUN 设备
	if vm.tunDevice != nil {
		vm.tunDevice.Close()
		vm.tunDevice = nil
	}

	vm.running = false
	return nil
}

// IsRunning 检查 VPN 是否运行
func (vm *VPNManager) IsRunning() bool {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.running
}

// GetStats 获取 VPN 统计信息
func (vm *VPNManager) GetStats() (*VPNStats, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	if !vm.running {
		return nil, fmt.Errorf("VPN is not running")
	}

	stats := &VPNStats{
		Running:   vm.running,
		SocksAddr: vm.config.SocksAddr,
		MTU:       vm.config.TunMTU,
	}

	return stats, nil
}

// validateConfig 验证 VPN 配置
func (vm *VPNManager) validateConfig(config *VPNConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if config.TunFd <= 0 {
		return fmt.Errorf("invalid TUN file descriptor: %d", config.TunFd)
	}

	if config.SocksAddr == "" {
		return fmt.Errorf("SOCKS address is required")
	}

	if config.TunMTU <= 0 {
		config.TunMTU = 1500 // 默认 MTU
	}

	if len(config.DNSServers) == 0 {
		// 使用默认 DNS
		config.DNSServers = []string{"8.8.8.8", "8.8.4.4"}
	}

	return nil
}

// VPNStats VPN 统计信息
type VPNStats struct {
	Running   bool   `json:"running"`
	SocksAddr string `json:"socksAddr"`
	MTU       int    `json:"mtu"`
}
