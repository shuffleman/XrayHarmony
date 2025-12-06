package main

import (
	"encoding/json"
	"fmt"

	"github.com/xtls/xray-core/infra/conf"
)

// ConfigBuilder 配置构建器
type ConfigBuilder struct {
	LogLevel   string
	Inbounds   []*conf.InboundDetourConfig
	Outbounds  []*conf.OutboundDetourConfig
	Routing    *conf.RouterConfig
	DNS        *conf.DNSConfig
	Policy     *conf.PolicyConfig
	Stats      *conf.StatsConfig
}

// NewConfigBuilder 创建新的配置构建器
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		LogLevel:  "warning",
		Inbounds:  make([]*conf.InboundDetourConfig, 0),
		Outbounds: make([]*conf.OutboundDetourConfig, 0),
	}
}

// SetLogLevel 设置日志级别
func (b *ConfigBuilder) SetLogLevel(level string) *ConfigBuilder {
	b.LogLevel = level
	return b
}

// AddSocksInbound 添加 SOCKS5 入站
func (b *ConfigBuilder) AddSocksInbound(port uint16, listen string, auth bool, udp bool) *ConfigBuilder {
	settings := &conf.SocksServerConfig{
		AuthType: "noauth",
		UDPEnabled: udp,
	}

	if auth {
		settings.AuthType = "password"
	}

	settingsMsg, _ := json.Marshal(settings)

	inbound := &conf.InboundDetourConfig{
		Tag:      "socks-in",
		Protocol: "socks",
		PortList: &conf.PortList{Range: []conf.PortRange{{From: port, To: port}}},
		ListenOn: &conf.Address{Address: &conf.IPOrDomain{IP: []byte(listen)}},
		Settings: (*json.RawMessage)(&settingsMsg),
	}

	b.Inbounds = append(b.Inbounds, inbound)
	return b
}

// AddHTTPInbound 添加 HTTP 入站
func (b *ConfigBuilder) AddHTTPInbound(port uint16, listen string) *ConfigBuilder {
	settings := &conf.HTTPServerConfig{
		Timeout: 300,
	}

	settingsMsg, _ := json.Marshal(settings)

	inbound := &conf.InboundDetourConfig{
		Tag:      "http-in",
		Protocol: "http",
		PortList: &conf.PortList{Range: []conf.PortRange{{From: port, To: port}}},
		ListenOn: &conf.Address{Address: &conf.IPOrDomain{IP: []byte(listen)}},
		Settings: (*json.RawMessage)(&settingsMsg),
	}

	b.Inbounds = append(b.Inbounds, inbound)
	return b
}

// AddVMessOutbound 添加 VMess 出站
func (b *ConfigBuilder) AddVMessOutbound(address string, port uint16, uuid string, alterID uint16, security string) *ConfigBuilder {
	vnext := []map[string]interface{}{
		{
			"address": address,
			"port":    port,
			"users": []map[string]interface{}{
				{
					"id":       uuid,
					"alterId":  alterID,
					"security": security,
				},
			},
		},
	}

	settings := map[string]interface{}{
		"vnext": vnext,
	}

	settingsMsg, _ := json.Marshal(settings)

	outbound := &conf.OutboundDetourConfig{
		Tag:      "proxy",
		Protocol: "vmess",
		Settings: (*json.RawMessage)(&settingsMsg),
	}

	b.Outbounds = append(b.Outbounds, outbound)
	return b
}

// AddVLESSOutbound 添加 VLESS 出站
func (b *ConfigBuilder) AddVLESSOutbound(address string, port uint16, uuid string, flow string, encryption string) *ConfigBuilder {
	vnext := []map[string]interface{}{
		{
			"address": address,
			"port":    port,
			"users": []map[string]interface{}{
				{
					"id":         uuid,
					"flow":       flow,
					"encryption": encryption,
				},
			},
		},
	}

	settings := map[string]interface{}{
		"vnext": vnext,
	}

	settingsMsg, _ := json.Marshal(settings)

	outbound := &conf.OutboundDetourConfig{
		Tag:      "proxy",
		Protocol: "vless",
		Settings: (*json.RawMessage)(&settingsMsg),
	}

	b.Outbounds = append(b.Outbounds, outbound)
	return b
}

// AddTrojanOutbound 添加 Trojan 出站
func (b *ConfigBuilder) AddTrojanOutbound(address string, port uint16, password string) *ConfigBuilder {
	servers := []map[string]interface{}{
		{
			"address":  address,
			"port":     port,
			"password": password,
		},
	}

	settings := map[string]interface{}{
		"servers": servers,
	}

	settingsMsg, _ := json.Marshal(settings)

	outbound := &conf.OutboundDetourConfig{
		Tag:      "proxy",
		Protocol: "trojan",
		Settings: (*json.RawMessage)(&settingsMsg),
	}

	b.Outbounds = append(b.Outbounds, outbound)
	return b
}

// AddShadowsocksOutbound 添加 Shadowsocks 出站
func (b *ConfigBuilder) AddShadowsocksOutbound(address string, port uint16, password string, method string) *ConfigBuilder {
	servers := []map[string]interface{}{
		{
			"address":  address,
			"port":     port,
			"password": password,
			"method":   method,
		},
	}

	settings := map[string]interface{}{
		"servers": servers,
	}

	settingsMsg, _ := json.Marshal(settings)

	outbound := &conf.OutboundDetourConfig{
		Tag:      "proxy",
		Protocol: "shadowsocks",
		Settings: (*json.RawMessage)(&settingsMsg),
	}

	b.Outbounds = append(b.Outbounds, outbound)
	return b
}

// AddFreedomOutbound 添加 Freedom 出站 (直连)
func (b *ConfigBuilder) AddFreedomOutbound(tag string) *ConfigBuilder {
	settings := map[string]interface{}{
		"domainStrategy": "UseIP",
	}

	settingsMsg, _ := json.Marshal(settings)

	outbound := &conf.OutboundDetourConfig{
		Tag:      tag,
		Protocol: "freedom",
		Settings: (*json.RawMessage)(&settingsMsg),
	}

	b.Outbounds = append(b.Outbounds, outbound)
	return b
}

// AddBlackholeOutbound 添加 Blackhole 出站 (拦截)
func (b *ConfigBuilder) AddBlackholeOutbound(tag string) *ConfigBuilder {
	outbound := &conf.OutboundDetourConfig{
		Tag:      tag,
		Protocol: "blackhole",
	}

	b.Outbounds = append(b.Outbounds, outbound)
	return b
}

// SetRouting 设置路由规则
func (b *ConfigBuilder) SetRouting(rules []*conf.RouterRule, domainStrategy string) *ConfigBuilder {
	b.Routing = &conf.RouterConfig{
		DomainStrategy: &domainStrategy,
		Rules:          rules,
	}
	return b
}

// AddRoutingRule 添加路由规则
func (b *ConfigBuilder) AddRoutingRule(ruleType string, outboundTag string, domains []string, ips []string) *ConfigBuilder {
	if b.Routing == nil {
		strategy := "AsIs"
		b.Routing = &conf.RouterConfig{
			DomainStrategy: &strategy,
			Rules:          make([]*conf.RouterRule, 0),
		}
	}

	rule := &conf.RouterRule{
		Type:        ruleType,
		OutboundTag: outboundTag,
	}

	if len(domains) > 0 {
		domainList := make([]*conf.StringList, len(domains))
		for i, d := range domains {
			domainList[i] = &conf.StringList{Value: d}
		}
		rule.Domain = domainList
	}

	if len(ips) > 0 {
		ipList := make([]*conf.StringList, len(ips))
		for i, ip := range ips {
			ipList[i] = &conf.StringList{Value: ip}
		}
		rule.IP = ipList
	}

	b.Routing.Rules = append(b.Routing.Rules, rule)
	return b
}

// SetDNS 设置 DNS
func (b *ConfigBuilder) SetDNS(servers []string, hosts map[string]string) *ConfigBuilder {
	dnsServers := make([]*conf.NameServerConfig, len(servers))
	for i, s := range servers {
		dnsServers[i] = &conf.NameServerConfig{
			Address: &conf.Address{Address: &conf.IPOrDomain{IP: []byte(s)}},
		}
	}

	dnsHosts := make(map[string]*conf.Address)
	for k, v := range hosts {
		dnsHosts[k] = &conf.Address{Address: &conf.IPOrDomain{IP: []byte(v)}}
	}

	b.DNS = &conf.DNSConfig{
		Servers: dnsServers,
		Hosts:   dnsHosts,
	}
	return b
}

// EnableStats 启用统计
func (b *ConfigBuilder) EnableStats() *ConfigBuilder {
	b.Stats = &conf.StatsConfig{}
	return b
}

// EnablePolicy 启用策略
func (b *ConfigBuilder) EnablePolicy(handshake, connIdle, uplinkOnly, downlinkOnly uint32) *ConfigBuilder {
	b.Policy = &conf.PolicyConfig{
		Levels: map[uint32]*conf.Policy{
			0: {
				Timeout: &conf.Policy_Timeout{
					Handshake:    &handshake,
					ConnectionIdle: &connIdle,
					UplinkOnly:   &uplinkOnly,
					DownlinkOnly: &downlinkOnly,
				},
			},
		},
	}
	return b
}

// Build 构建配置
func (b *ConfigBuilder) Build() (*conf.Config, error) {
	if len(b.Inbounds) == 0 {
		return nil, fmt.Errorf("at least one inbound is required")
	}

	if len(b.Outbounds) == 0 {
		return nil, fmt.Errorf("at least one outbound is required")
	}

	config := &conf.Config{
		LogConfig: &conf.LogConfig{
			LogLevel: b.LogLevel,
		},
		InboundConfigs:  b.Inbounds,
		OutboundConfigs: b.Outbounds,
	}

	if b.Routing != nil {
		config.RouterConfig = b.Routing
	}

	if b.DNS != nil {
		config.DNSConfig = b.DNS
	}

	if b.Policy != nil {
		config.Policy = b.Policy
	}

	if b.Stats != nil {
		config.Stats = b.Stats
	}

	return config, nil
}

// BuildJSON 构建 JSON 配置
func (b *ConfigBuilder) BuildJSON() (string, error) {
	config, err := b.Build()
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	return string(data), nil
}
