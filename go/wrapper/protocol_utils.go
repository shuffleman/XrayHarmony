package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	Protocol  string            `json:"protocol"`
	Address   string            `json:"address"`
	Port      uint16            `json:"port"`
	ID        string            `json:"id,omitempty"`        // UUID for VMess/VLESS
	AlterID   uint16            `json:"alterId,omitempty"`   // VMess AlterID
	Security  string            `json:"security,omitempty"`  // VMess security
	Encryption string           `json:"encryption,omitempty"` // VLESS encryption
	Flow      string            `json:"flow,omitempty"`      // VLESS flow
	Password  string            `json:"password,omitempty"`  // Trojan/Shadowsocks password
	Method    string            `json:"method,omitempty"`    // Shadowsocks method
	Network   string            `json:"network,omitempty"`   // transport (tcp/ws/grpc/h2)
	Type      string            `json:"type,omitempty"`      // header type
	Host      string            `json:"host,omitempty"`      // WS/H2 host
	Path      string            `json:"path,omitempty"`      // WS/H2/gRPC path
	TLS       string            `json:"tls,omitempty"`       // tls/xtls/reality
	SNI       string            `json:"sni,omitempty"`       // TLS server name
	ALPN      string            `json:"alpn,omitempty"`      // ALPN
	Fingerprint string          `json:"fp,omitempty"`        // TLS fingerprint
	PublicKey string            `json:"pbk,omitempty"`       // Reality public key
	ShortID   string            `json:"sid,omitempty"`       // Reality short ID
	SpiderX   string            `json:"spx,omitempty"`       // Reality spider X
	Remark    string            `json:"remark,omitempty"`    // 备注名称
}

// ParseVMessURL 解析 VMess URL
// 格式: vmess://base64(json)
func ParseVMessURL(vmessURL string) (*ServerConfig, error) {
	// 移除 vmess:// 前缀
	if !strings.HasPrefix(vmessURL, "vmess://") {
		return nil, fmt.Errorf("invalid vmess URL scheme")
	}

	encoded := strings.TrimPrefix(vmessURL, "vmess://")

	// Base64 解码
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		// 尝试 RawStdEncoding
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("failed to decode vmess URL: %w", err)
		}
	}

	// 解析 JSON
	var vmessConfig struct {
		V    string `json:"v"`
		PS   string `json:"ps"`
		Add  string `json:"add"`
		Port string `json:"port"`
		ID   string `json:"id"`
		Aid  string `json:"aid"`
		Scy  string `json:"scy"`
		Net  string `json:"net"`
		Type string `json:"type"`
		Host string `json:"host"`
		Path string `json:"path"`
		TLS  string `json:"tls"`
		SNI  string `json:"sni"`
		ALPN string `json:"alpn"`
	}

	if err := json.Unmarshal(decoded, &vmessConfig); err != nil {
		return nil, fmt.Errorf("failed to parse vmess JSON: %w", err)
	}

	port, _ := strconv.ParseUint(vmessConfig.Port, 10, 16)
	alterID, _ := strconv.ParseUint(vmessConfig.Aid, 10, 16)

	config := &ServerConfig{
		Protocol: "vmess",
		Address:  vmessConfig.Add,
		Port:     uint16(port),
		ID:       vmessConfig.ID,
		AlterID:  uint16(alterID),
		Security: vmessConfig.Scy,
		Network:  vmessConfig.Net,
		Type:     vmessConfig.Type,
		Host:     vmessConfig.Host,
		Path:     vmessConfig.Path,
		TLS:      vmessConfig.TLS,
		SNI:      vmessConfig.SNI,
		ALPN:     vmessConfig.ALPN,
		Remark:   vmessConfig.PS,
	}

	return config, nil
}

// ParseVLESSURL 解析 VLESS URL
// 格式: vless://uuid@address:port?encryption=none&flow=xtls-rprx-vision&security=tls&sni=example.com&type=tcp#remark
func ParseVLESSURL(vlessURL string) (*ServerConfig, error) {
	if !strings.HasPrefix(vlessURL, "vless://") {
		return nil, fmt.Errorf("invalid vless URL scheme")
	}

	// 移除 vless:// 前缀
	urlStr := strings.TrimPrefix(vlessURL, "vless://")

	// 分离备注
	var remark string
	if idx := strings.Index(urlStr, "#"); idx != -1 {
		remark, _ = url.QueryUnescape(urlStr[idx+1:])
		urlStr = urlStr[:idx]
	}

	// 分离参数
	parts := strings.SplitN(urlStr, "?", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid vless URL format")
	}

	// 解析 uuid@address:port
	userAndAddr := parts[0]
	atIdx := strings.LastIndex(userAndAddr, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid vless URL format: missing @")
	}

	uuid := userAndAddr[:atIdx]
	addrPort := userAndAddr[atIdx+1:]

	colonIdx := strings.LastIndex(addrPort, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid vless URL format: missing port")
	}

	address := addrPort[:colonIdx]
	portStr := addrPort[colonIdx+1:]
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	// 解析查询参数
	params, err := url.ParseQuery(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse query params: %w", err)
	}

	config := &ServerConfig{
		Protocol:   "vless",
		Address:    address,
		Port:       uint16(port),
		ID:         uuid,
		Encryption: params.Get("encryption"),
		Flow:       params.Get("flow"),
		TLS:        params.Get("security"),
		SNI:        params.Get("sni"),
		ALPN:       params.Get("alpn"),
		Network:    params.Get("type"),
		Host:       params.Get("host"),
		Path:       params.Get("path"),
		Fingerprint: params.Get("fp"),
		PublicKey:  params.Get("pbk"),
		ShortID:    params.Get("sid"),
		SpiderX:    params.Get("spx"),
		Remark:     remark,
	}

	return config, nil
}

// ParseTrojanURL 解析 Trojan URL
// 格式: trojan://password@address:port?security=tls&sni=example.com&type=tcp#remark
func ParseTrojanURL(trojanURL string) (*ServerConfig, error) {
	if !strings.HasPrefix(trojanURL, "trojan://") {
		return nil, fmt.Errorf("invalid trojan URL scheme")
	}

	urlStr := strings.TrimPrefix(trojanURL, "trojan://")

	// 分离备注
	var remark string
	if idx := strings.Index(urlStr, "#"); idx != -1 {
		remark, _ = url.QueryUnescape(urlStr[idx+1:])
		urlStr = urlStr[:idx]
	}

	// 分离参数
	parts := strings.SplitN(urlStr, "?", 2)

	// 解析 password@address:port
	userAndAddr := parts[0]
	atIdx := strings.LastIndex(userAndAddr, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid trojan URL format: missing @")
	}

	password := userAndAddr[:atIdx]
	addrPort := userAndAddr[atIdx+1:]

	colonIdx := strings.LastIndex(addrPort, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid trojan URL format: missing port")
	}

	address := addrPort[:colonIdx]
	portStr := addrPort[colonIdx+1:]
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	config := &ServerConfig{
		Protocol: "trojan",
		Address:  address,
		Port:     uint16(port),
		Password: password,
		Remark:   remark,
	}

	// 解析查询参数 (如果有)
	if len(parts) == 2 {
		params, err := url.ParseQuery(parts[1])
		if err == nil {
			config.TLS = params.Get("security")
			config.SNI = params.Get("sni")
			config.ALPN = params.Get("alpn")
			config.Network = params.Get("type")
			config.Host = params.Get("host")
			config.Path = params.Get("path")
			config.Fingerprint = params.Get("fp")
		}
	}

	return config, nil
}

// ParseShadowsocksURL 解析 Shadowsocks URL
// 格式: ss://base64(method:password)@address:port#remark
func ParseShadowsocksURL(ssURL string) (*ServerConfig, error) {
	if !strings.HasPrefix(ssURL, "ss://") {
		return nil, fmt.Errorf("invalid shadowsocks URL scheme")
	}

	urlStr := strings.TrimPrefix(ssURL, "ss://")

	// 分离备注
	var remark string
	if idx := strings.Index(urlStr, "#"); idx != -1 {
		remark, _ = url.QueryUnescape(urlStr[idx+1:])
		urlStr = urlStr[:idx]
	}

	// 分离 @ 符号
	atIdx := strings.LastIndex(urlStr, "@")
	if atIdx == -1 {
		return nil, fmt.Errorf("invalid shadowsocks URL format: missing @")
	}

	// Base64 解码 method:password
	encoded := urlStr[:atIdx]
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(encoded)
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(encoded)
			if err != nil {
				decoded, err = base64.RawURLEncoding.DecodeString(encoded)
				if err != nil {
					return nil, fmt.Errorf("failed to decode shadowsocks credentials: %w", err)
				}
			}
		}
	}

	credentials := string(decoded)
	colonIdx := strings.Index(credentials, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid shadowsocks credentials format")
	}

	method := credentials[:colonIdx]
	password := credentials[colonIdx+1:]

	// 解析 address:port
	addrPort := urlStr[atIdx+1:]
	colonIdx = strings.LastIndex(addrPort, ":")
	if colonIdx == -1 {
		return nil, fmt.Errorf("invalid shadowsocks URL format: missing port")
	}

	address := addrPort[:colonIdx]
	portStr := addrPort[colonIdx+1:]
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid port: %w", err)
	}

	config := &ServerConfig{
		Protocol: "shadowsocks",
		Address:  address,
		Port:     uint16(port),
		Method:   method,
		Password: password,
		Remark:   remark,
	}

	return config, nil
}

// GenerateVMessURL 生成 VMess URL
func GenerateVMessURL(config *ServerConfig) (string, error) {
	vmessConfig := map[string]interface{}{
		"v":    "2",
		"ps":   config.Remark,
		"add":  config.Address,
		"port": strconv.FormatUint(uint64(config.Port), 10),
		"id":   config.ID,
		"aid":  strconv.FormatUint(uint64(config.AlterID), 10),
		"scy":  config.Security,
		"net":  config.Network,
		"type": config.Type,
		"host": config.Host,
		"path": config.Path,
		"tls":  config.TLS,
		"sni":  config.SNI,
		"alpn": config.ALPN,
	}

	jsonData, err := json.Marshal(vmessConfig)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	return "vmess://" + encoded, nil
}

// GenerateVLESSURL 生成 VLESS URL
func GenerateVLESSURL(config *ServerConfig) string {
	params := url.Values{}
	if config.Encryption != "" {
		params.Set("encryption", config.Encryption)
	}
	if config.Flow != "" {
		params.Set("flow", config.Flow)
	}
	if config.TLS != "" {
		params.Set("security", config.TLS)
	}
	if config.SNI != "" {
		params.Set("sni", config.SNI)
	}
	if config.ALPN != "" {
		params.Set("alpn", config.ALPN)
	}
	if config.Network != "" {
		params.Set("type", config.Network)
	}
	if config.Host != "" {
		params.Set("host", config.Host)
	}
	if config.Path != "" {
		params.Set("path", config.Path)
	}
	if config.Fingerprint != "" {
		params.Set("fp", config.Fingerprint)
	}
	if config.PublicKey != "" {
		params.Set("pbk", config.PublicKey)
	}
	if config.ShortID != "" {
		params.Set("sid", config.ShortID)
	}
	if config.SpiderX != "" {
		params.Set("spx", config.SpiderX)
	}

	urlStr := fmt.Sprintf("vless://%s@%s:%d?%s",
		config.ID,
		config.Address,
		config.Port,
		params.Encode(),
	)

	if config.Remark != "" {
		urlStr += "#" + url.QueryEscape(config.Remark)
	}

	return urlStr
}

// GenerateTrojanURL 生成 Trojan URL
func GenerateTrojanURL(config *ServerConfig) string {
	params := url.Values{}
	if config.TLS != "" {
		params.Set("security", config.TLS)
	}
	if config.SNI != "" {
		params.Set("sni", config.SNI)
	}
	if config.ALPN != "" {
		params.Set("alpn", config.ALPN)
	}
	if config.Network != "" {
		params.Set("type", config.Network)
	}
	if config.Host != "" {
		params.Set("host", config.Host)
	}
	if config.Path != "" {
		params.Set("path", config.Path)
	}
	if config.Fingerprint != "" {
		params.Set("fp", config.Fingerprint)
	}

	urlStr := fmt.Sprintf("trojan://%s@%s:%d",
		config.Password,
		config.Address,
		config.Port,
	)

	if len(params) > 0 {
		urlStr += "?" + params.Encode()
	}

	if config.Remark != "" {
		urlStr += "#" + url.QueryEscape(config.Remark)
	}

	return urlStr
}

// GenerateShadowsocksURL 生成 Shadowsocks URL
func GenerateShadowsocksURL(config *ServerConfig) string {
	credentials := fmt.Sprintf("%s:%s", config.Method, config.Password)
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))

	urlStr := fmt.Sprintf("ss://%s@%s:%d",
		encoded,
		config.Address,
		config.Port,
	)

	if config.Remark != "" {
		urlStr += "#" + url.QueryEscape(config.Remark)
	}

	return urlStr
}

// ParseShareURL 解析分享链接 (自动识别类型)
func ParseShareURL(shareURL string) (*ServerConfig, error) {
	switch {
	case strings.HasPrefix(shareURL, "vmess://"):
		return ParseVMessURL(shareURL)
	case strings.HasPrefix(shareURL, "vless://"):
		return ParseVLESSURL(shareURL)
	case strings.HasPrefix(shareURL, "trojan://"):
		return ParseTrojanURL(shareURL)
	case strings.HasPrefix(shareURL, "ss://"):
		return ParseShadowsocksURL(shareURL)
	default:
		return nil, fmt.Errorf("unsupported share URL scheme")
	}
}

// GenerateShareURL 生成分享链接
func GenerateShareURL(config *ServerConfig) (string, error) {
	switch config.Protocol {
	case "vmess":
		return GenerateVMessURL(config)
	case "vless":
		return GenerateVLESSURL(config), nil
	case "trojan":
		return GenerateTrojanURL(config), nil
	case "shadowsocks":
		return GenerateShadowsocksURL(config), nil
	default:
		return "", fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}
}
