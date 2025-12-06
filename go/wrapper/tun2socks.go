package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Tun2SocksInstance tun2socks 实例
type Tun2SocksInstance struct {
	tunFd        int
	socksAddr    string
	mtu          int
	dnsAddr      string
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	running      bool
	bytesUp      uint64
	bytesDown    uint64
	statsLock    sync.RWMutex
}

// Tun2SocksConfig tun2socks 配置
type Tun2SocksConfig struct {
	TunFd     int    `json:"tun_fd"`      // TUN 设备文件描述符
	SocksAddr string `json:"socks_addr"`  // SOCKS5 代理地址，如 "127.0.0.1:10808"
	MTU       int    `json:"mtu"`         // MTU 大小，默认 1500
	DNSAddr   string `json:"dns_addr"`    // DNS 服务器地址，如 "8.8.8.8:53"
	FakeDNS   bool   `json:"fake_dns"`    // 是否启用 FakeDNS
}

// NewTun2SocksInstance 创建新的 tun2socks 实例
func NewTun2SocksInstance(config *Tun2SocksConfig) *Tun2SocksInstance {
	if config.MTU == 0 {
		config.MTU = 1500
	}

	if config.DNSAddr == "" {
		config.DNSAddr = "8.8.8.8:53"
	}

	return &Tun2SocksInstance{
		tunFd:     config.TunFd,
		socksAddr: config.SocksAddr,
		mtu:       config.MTU,
		dnsAddr:   config.DNSAddr,
		running:   false,
	}
}

// Start 启动 tun2socks
func (t *Tun2SocksInstance) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return fmt.Errorf("tun2socks is already running")
	}

	if t.tunFd < 0 {
		return fmt.Errorf("invalid TUN file descriptor")
	}

	if t.socksAddr == "" {
		return fmt.Errorf("SOCKS5 address is required")
	}

	// 验证 SOCKS5 代理是否可用
	if err := t.checkSocksProxy(); err != nil {
		return fmt.Errorf("SOCKS5 proxy not available: %w", err)
	}

	t.ctx, t.cancel = context.WithCancel(context.Background())

	// 注意: 实际的 tun2socks 实现需要使用专门的库如 gvisor 或 lwIP
	// 这里提供一个框架,实际使用时需要集成真实的 tun2socks 库
	// 推荐使用: github.com/xjasonlyu/tun2socks/v2

	go t.runTun2Socks()

	t.running = true
	return nil
}

// Stop 停止 tun2socks
func (t *Tun2SocksInstance) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.running {
		return fmt.Errorf("tun2socks is not running")
	}

	if t.cancel != nil {
		t.cancel()
	}

	t.running = false
	return nil
}

// IsRunning 检查是否正在运行
func (t *Tun2SocksInstance) IsRunning() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.running
}

// GetStats 获取流量统计
func (t *Tun2SocksInstance) GetStats() (bytesUp, bytesDown uint64) {
	t.statsLock.RLock()
	defer t.statsLock.RUnlock()
	return t.bytesUp, t.bytesDown
}

// ResetStats 重置统计
func (t *Tun2SocksInstance) ResetStats() {
	t.statsLock.Lock()
	defer t.statsLock.Unlock()
	t.bytesUp = 0
	t.bytesDown = 0
}

// checkSocksProxy 检查 SOCKS5 代理是否可用
func (t *Tun2SocksInstance) checkSocksProxy() error {
	conn, err := net.DialTimeout("tcp", t.socksAddr, 3*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// runTun2Socks 运行 tun2socks 主循环
func (t *Tun2SocksInstance) runTun2Socks() {
	// 这里是 tun2socks 的主要逻辑
	// 实际实现需要:
	// 1. 从 TUN 设备读取 IP 数据包
	// 2. 解析数据包，提取目标地址和端口
	// 3. 通过 SOCKS5 代理建立连接
	// 4. 转发数据并更新统计信息
	//
	// 推荐使用现有的 tun2socks 库:
	// - github.com/xjasonlyu/tun2socks/v2 (推荐)
	// - gvisor.dev/gvisor/pkg/tcpip (Google gVisor)
	//
	// 示例伪代码:
	/*
		engine, err := core.NewEngine(core.Config{
			TUN:        t.tunFd,
			Proxy:      t.socksAddr,
			MTU:        t.mtu,
			DNSServer:  t.dnsAddr,
		})
		if err != nil {
			return
		}

		<-t.ctx.Done()
		engine.Close()
	*/

	<-t.ctx.Done()
}

// addBytesUp 增加上行流量
func (t *Tun2SocksInstance) addBytesUp(n uint64) {
	t.statsLock.Lock()
	defer t.statsLock.Unlock()
	t.bytesUp += n
}

// addBytesDown 增加下行流量
func (t *Tun2SocksInstance) addBytesDown(n uint64) {
	t.statsLock.Lock()
	defer t.statsLock.Unlock()
	t.bytesDown += n
}

// forwardTraffic 转发流量 (辅助函数)
func (t *Tun2SocksInstance) forwardTraffic(src, dst io.ReadWriter, direction string) {
	buf := make([]byte, t.mtu)
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			n, err := src.Read(buf)
			if err != nil {
				return
			}

			if n > 0 {
				if _, err := dst.Write(buf[:n]); err != nil {
					return
				}

				// 更新统计
				if direction == "up" {
					t.addBytesUp(uint64(n))
				} else {
					t.addBytesDown(uint64(n))
				}
			}
		}
	}
}

// ===============================
// TUN 设备辅助函数
// ===============================

// TUNConfig TUN 设备配置
type TUNConfig struct {
	Name    string   `json:"name"`     // 设备名称，如 "tun0"
	Address string   `json:"address"`  // IP 地址，如 "10.0.0.2"
	Gateway string   `json:"gateway"`  // 网关，如 "10.0.0.1"
	Netmask string   `json:"netmask"`  // 子网掩码，如 "255.255.255.0"
	MTU     int      `json:"mtu"`      // MTU 大小
	DNS     []string `json:"dns"`      // DNS 服务器列表
	Routes  []string `json:"routes"`   // 路由规则列表
}

// 注意: HarmonyOS 的 TUN 设备创建和配置通过系统 VPN API 完成
// 不需要在 Go 层面直接操作 TUN 设备
// 以下函数仅作为参考和说明

// CreateTUNDevice 创建 TUN 设备 (仅供参考，实际由 HarmonyOS VPN API 处理)
func CreateTUNDevice(config *TUNConfig) (int, error) {
	// 在 HarmonyOS 中，TUN 设备由 VPN 扩展能力 (VPNExtensionAbility) 创建
	// 应用层通过 @ohos.net.vpnExtension API 配置和获取 TUN 设备的文件描述符
	//
	// 典型流程:
	// 1. 在 VPNExtensionAbility 中调用 setUp() 方法
	// 2. 配置 VPN 参数 (地址、路由、DNS 等)
	// 3. 建立 VPN 连接，获取 TUN 文件描述符
	// 4. 将文件描述符传递给 tun2socks

	return -1, fmt.Errorf("TUN device should be created via HarmonyOS VPN API")
}

// ConfigureTUNDevice 配置 TUN 设备 (仅供参考)
func ConfigureTUNDevice(fd int, config *TUNConfig) error {
	// HarmonyOS VPN API 在创建 TUN 设备时已经完成配置
	// 包括: IP 地址、路由、DNS、MTU 等

	return fmt.Errorf("TUN device should be configured via HarmonyOS VPN API")
}
