# XrayHarmony 完整 API 文档

本文档描述 XrayHarmony 为 HarmonyOS 提供的完整 Xray 封装 API。

## 目录

- [核心功能](#核心功能)
- [协议工具](#协议工具)
- [Tun2Socks](#tun2socks)
- [资产管理](#资产管理)
- [配置构建器](#配置构建器)
- [使用示例](#使用示例)

## 核心功能

### Xray 实例管理

XrayHarmony 提供完整的 Xray-core 生命周期管理。

#### 创建实例

```go
// Go 层 API
instance := NewXrayInstance()
instance := NewXrayInstanceWithAssets(assetDir string)
```

```c
// C API
long long XrayNewInstance();
```

#### 加载配置

```go
// 从 JSON 字符串加载
err := instance.LoadConfig(configJSON string)

// 从文件加载
err := instance.LoadConfigFromFile(filePath string)
```

```c
// C API
int XrayLoadConfig(long long id, const char* configJSON);
int XrayLoadConfigFromFile(long long id, const char* filePath);
```

#### 启动/停止

```go
err := instance.Start()
err := instance.Stop()
running := instance.IsRunning()
```

```c
int XrayStart(long long id);
int XrayStop(long long id);
int XrayIsRunning(long long id);  // 返回 1=运行中, 0=已停止, -1=错误
```

#### 获取统计信息

```go
statsJSON, err := instance.GetStats()
// 返回 JSON:
// {
//   "running": true,
//   "status": "ok",
//   "uptime": 123.45,
//   "traffic": {
//     "uplink": 1024000,
//     "downlink": 2048000
//   }
// }
```

```c
char* XrayGetStats(long long id);
```

#### 测试配置

```go
err := instance.TestConfig(configJSON string)
```

```c
int XrayTestConfig(long long id, const char* configJSON);
```

#### 删除实例

```go
// 实例会在 GC 时自动清理
```

```c
int XrayDeleteInstance(long long id);
```

## 协议工具

XrayHarmony 提供完整的协议解析和生成工具,支持:
- VMess (vmess://)
- VLESS (vless://)
- Trojan (trojan://)
- Shadowsocks (ss://)

### 解析分享链接

```go
// 自动识别协议类型并解析
config, err := ParseShareURL(shareURL string)

// 或使用特定协议函数
vmessConfig, err := ParseVMessURL(vmessURL string)
vlessConfig, err := ParseVLESSURL(vlessURL string)
trojanConfig, err := ParseTrojanURL(trojanURL string)
ssConfig, err := ParseShadowsocksURL(ssURL string)

// 返回的 ServerConfig 结构:
type ServerConfig struct {
    Protocol    string  // vmess, vless, trojan, shadowsocks
    Address     string  // 服务器地址
    Port        uint16  // 端口
    ID          string  // UUID (VMess/VLESS)
    AlterID     uint16  // VMess AlterID
    Security    string  // 加密方式
    Password    string  // 密码 (Trojan/SS)
    Method      string  // 加密方法 (SS)
    Network     string  // 传输协议 (tcp/ws/grpc/h2)
    TLS         string  // tls/xtls/reality
    // ... 更多字段
}
```

```c
// C API
char* XrayParseShareURL(const char* shareURL);
// 返回 ServerConfig 的 JSON 表示
```

### 生成分享链接

```go
// 从 ServerConfig 生成分享链接
shareURL, err := GenerateShareURL(config *ServerConfig)

// 或使用特定协议函数
vmessURL, err := GenerateVMessURL(config *ServerConfig)
vlessURL := GenerateVLESSURL(config *ServerConfig)
trojanURL := GenerateTrojanURL(config *ServerConfig)
ssURL := GenerateShadowsocksURL(config *ServerConfig)
```

```c
// C API
char* XrayGenerateShareURL(const char* configJSON);
```

## Tun2Socks

Tun2Socks 用于将 TUN 设备的流量转发到 SOCKS5 代理,实现 VPN 功能。

### 创建实例

```go
config := &Tun2SocksConfig{
    TunFd:     tunFileDescriptor,  // TUN 设备 FD
    SocksAddr: "127.0.0.1:10808",  // SOCKS5 地址
    MTU:       1500,                // MTU 大小
    DNSAddr:   "8.8.8.8:53",       // DNS 服务器
    FakeDNS:   false,               // 是否启用 FakeDNS
}

instance := NewTun2SocksInstance(config)
```

```c
// C API
long long Tun2SocksNew(const char* configJSON);
```

### 启动/停止

```go
err := instance.Start()
err := instance.Stop()
running := instance.IsRunning()
```

```c
int Tun2SocksStart(long long id);
int Tun2SocksStop(long long id);
int Tun2SocksIsRunning(long long id);
```

### 获取统计

```go
bytesUp, bytesDown := instance.GetStats()
instance.ResetStats()
```

```c
char* Tun2SocksGetStats(long long id);
// 返回: {"bytes_up": 1024, "bytes_down": 2048}
```

### 删除实例

```c
int Tun2SocksDelete(long long id);
```

### 注意事项

1. **TUN 设备**: 在 HarmonyOS 中,TUN 设备由 VPN Extension Ability API 创建
2. **文件描述符**: 需要将 TUN FD 传递给 tun2socks
3. **SOCKS5 代理**: 通常由 Xray 提供 SOCKS5 入站
4. **架构**: HarmonyOS VPN API → TUN → tun2socks → SOCKS5 → Xray

## 资产管理

管理 geoip.dat 和 geosite.dat 文件,用于路由规则。

### 创建管理器

```go
manager := NewAssetManager(baseDir string)
```

```c
long long AssetManagerNew(const char* baseDir);
```

### 获取资产信息

```go
info, err := manager.GetAssetInfo(AssetTypeGeoIP)
// 或
info, err := manager.GetAssetInfo(AssetTypeGeoSite)

// AssetInfo 结构:
type AssetInfo struct {
    Type         AssetType
    Version      string
    Size         int64
    LastModified time.Time
    Path         string
    Exists       bool
}
```

```c
char* AssetManagerGetInfo(long long id, const char* assetType);
// assetType: "geoip" 或 "geosite"
```

### 下载资产

```go
// 使用默认 URL (Loyalsoldier/v2ray-rules-dat)
err := manager.DownloadAsset(AssetTypeGeoIP, "", nil)

// 使用自定义 URL 和进度回调
err := manager.DownloadAsset(
    AssetTypeGeoIP,
    "https://example.com/geoip.dat",
    func(progress *DownloadProgress) {
        fmt.Printf("下载进度: %.2f%%\n", progress.Percentage)
    },
)
```

```c
int AssetManagerDownload(long long id, const char* assetType, const char* url);
// url 为空字符串时使用默认 URL
```

### 检查更新

```go
needsUpdate, err := manager.CheckAssetUpdate(AssetTypeGeoIP, "")
```

```c
int AssetManagerCheckUpdate(long long id, const char* assetType, const char* url);
// 返回 1=需要更新, 0=不需要, -1=错误
```

### 验证资产

```go
valid, err := manager.VerifyAsset(AssetTypeGeoIP)
```

```c
int AssetManagerVerify(long long id, const char* assetType);
// 返回 1=有效, 0=无效, -1=错误
```

### 获取所有资产

```go
assets, err := manager.GetAllAssets()
```

### 删除管理器

```c
int AssetManagerDelete(long long id);
```

## 配置构建器

提供流畅的 API 来构建 Xray 配置。

### 创建构建器

```go
builder := NewConfigBuilder()
```

```c
long long ConfigBuilderNew();
```

### 设置日志级别

```go
builder.SetLogLevel("warning")  // debug, info, warning, error, none
```

```c
int ConfigBuilderSetLogLevel(long long id, const char* level);
```

### 添加入站

```go
// SOCKS5 入站
builder.AddSocksInbound(port uint16, listen string, auth bool, udp bool)

// HTTP 入站
builder.AddHTTPInbound(port uint16, listen string)
```

### 添加出站

```go
// VMess 出站
builder.AddVMessOutbound(
    address string,
    port uint16,
    uuid string,
    alterID uint16,
    security string,  // auto, aes-128-gcm, chacha20-poly1305, none
)

// VLESS 出站
builder.AddVLESSOutbound(
    address string,
    port uint16,
    uuid string,
    flow string,        // xtls-rprx-vision, etc.
    encryption string,  // none
)

// Trojan 出站
builder.AddTrojanOutbound(
    address string,
    port uint16,
    password string,
)

// Shadowsocks 出站
builder.AddShadowsocksOutbound(
    address string,
    port uint16,
    password string,
    method string,  // aes-256-gcm, chacha20-poly1305, etc.
)

// Freedom 出站 (直连)
builder.AddFreedomOutbound(tag string)

// Blackhole 出站 (拦截)
builder.AddBlackholeOutbound(tag string)
```

### 设置路由

```go
// 添加路由规则
builder.AddRoutingRule(
    ruleType string,      // "field"
    outboundTag string,   // 出站标签
    domains []string,     // 域名列表
    ips []string,         // IP 列表
)
```

### 设置 DNS

```go
builder.SetDNS(
    servers []string,              // ["8.8.8.8", "1.1.1.1"]
    hosts map[string]string,       // {"example.com": "127.0.0.1"}
)
```

### 启用功能

```go
// 启用统计
builder.EnableStats()

// 启用策略
builder.EnablePolicy(
    handshake uint32,      // 握手超时 (秒)
    connIdle uint32,       // 连接空闲超时
    uplinkOnly uint32,     // 仅上行超时
    downlinkOnly uint32,   // 仅下行超时
)
```

### 构建配置

```go
// 构建为 Config 对象
config, err := builder.Build()

// 构建为 JSON 字符串
jsonStr, err := builder.BuildJSON()
```

```c
char* ConfigBuilderBuild(long long id);
// 返回 JSON 字符串
```

### 删除构建器

```c
int ConfigBuilderDelete(long long id);
```

### 完整示例

```go
builder := NewConfigBuilder()

// 设置日志
builder.SetLogLevel("warning")

// 添加 SOCKS5 入站
builder.AddSocksInbound(10808, "127.0.0.1", false, true)

// 添加 VMess 出站
builder.AddVMessOutbound(
    "example.com",
    443,
    "uuid-here",
    0,
    "auto",
)

// 添加直连出站
builder.AddFreedomOutbound("direct")

// 添加拦截出站
builder.AddBlackholeOutbound("block")

// 添加路由规则
builder.AddRoutingRule(
    "field",
    "direct",
    []string{"geosite:cn"},
    []string{"geoip:cn", "geoip:private"},
)

// 设置 DNS
builder.SetDNS(
    []string{"8.8.8.8", "1.1.1.1"},
    map[string]string{},
)

// 启用统计
builder.EnableStats()

// 构建配置
configJSON, err := builder.BuildJSON()
```

## 使用示例

### 基础 Xray 代理

```go
// 1. 创建实例
instance := NewXrayInstance()

// 2. 解析分享链接
serverConfig, _ := ParseShareURL("vmess://...")

// 3. 使用配置构建器
builder := NewConfigBuilder()
builder.SetLogLevel("warning")
builder.AddSocksInbound(10808, "127.0.0.1", false, true)
builder.AddVMessOutbound(
    serverConfig.Address,
    serverConfig.Port,
    serverConfig.ID,
    serverConfig.AlterID,
    serverConfig.Security,
)
configJSON, _ := builder.BuildJSON()

// 4. 加载配置并启动
instance.LoadConfig(configJSON)
instance.Start()

// 5. 检查状态
if instance.IsRunning() {
    stats, _ := instance.GetStats()
    fmt.Println(stats)
}

// 6. 停止
instance.Stop()
```

### VPN 模式 (Xray + Tun2Socks)

```go
// 1. 启动 Xray SOCKS5 代理
xrayInstance := NewXrayInstance()
builder := NewConfigBuilder()
builder.AddSocksInbound(10808, "127.0.0.1", false, true)
builder.AddVMessOutbound("server.com", 443, "uuid", 0, "auto")
xrayInstance.LoadConfig(builder.BuildJSON())
xrayInstance.Start()

// 2. 获取 TUN 设备 FD (通过 HarmonyOS VPN API)
tunFd := setupVPN()  // 你的 VPN 设置代码

// 3. 启动 Tun2Socks
tun2socksConfig := &Tun2SocksConfig{
    TunFd:     tunFd,
    SocksAddr: "127.0.0.1:10808",
    MTU:       1500,
    DNSAddr:   "8.8.8.8:53",
}
tun2socksInstance := NewTun2SocksInstance(tun2socksConfig)
tun2socksInstance.Start()

// 现在系统流量会经过: TUN → tun2socks → Xray SOCKS5 → 代理服务器
```

### 资产管理

```go
// 1. 创建资产管理器
assetMgr := NewAssetManager("/data/assets")

// 2. 检查并下载 geoip
needsUpdate, _ := assetMgr.CheckAssetUpdate(AssetTypeGeoIP, "")
if needsUpdate {
    assetMgr.DownloadAsset(AssetTypeGeoIP, "", func(p *DownloadProgress) {
        fmt.Printf("下载 geoip: %.2f%%\n", p.Percentage)
    })
}

// 3. 检查并下载 geosite
needsUpdate, _ = assetMgr.CheckAssetUpdate(AssetTypeGeoSite, "")
if needsUpdate {
    assetMgr.DownloadAsset(AssetTypeGeoSite, "", nil)
}

// 4. 创建带资产管理的 Xray 实例
instance := NewXrayInstanceWithAssets("/data/assets")

// 5. 在配置中使用 geoip/geosite
builder := NewConfigBuilder()
// ... 添加入站出站 ...
builder.AddRoutingRule("field", "direct",
    []string{"geosite:cn"},
    []string{"geoip:cn", "geoip:private"},
)
instance.LoadConfig(builder.BuildJSON())
instance.Start()
```

## 错误处理

所有 C API 函数都会在发生错误时设置错误信息,可以通过以下函数获取:

```c
char* XrayGetLastError();
void XrayFreeString(char* str);
```

使用示例:

```c
if (XrayStart(id) != 0) {
    char* error = XrayGetLastError();
    printf("错误: %s\n", error);
    XrayFreeString(error);
}
```

## 支持的协议

### 入站协议
- SOCKS5
- HTTP

### 出站协议
- VMess
- VLESS
- Trojan
- Shadowsocks
- Freedom (直连)
- Blackhole (拦截)

### 传输协议
- TCP
- WebSocket
- HTTP/2
- gRPC
- QUIC

### 安全协议
- TLS
- XTLS
- Reality

## 性能建议

1. **复用实例**: 尽量复用 Xray 实例而不是频繁创建/销毁
2. **资产缓存**: geoip/geosite 文件下载后会缓存,不需要每次都下载
3. **统计开销**: 只在需要时启用统计功能
4. **日志级别**: 生产环境建议使用 "warning" 或 "error"
5. **配置验证**: 使用 TestConfig 在应用配置前验证

## 线程安全

所有 Go 层 API 都是线程安全的,内部使用互斥锁保护。C API 通过 Go 层保证线程安全。

## 参考资料

- [Xray-core 文档](https://xtls.github.io/)
- [v2rayNG 项目](https://github.com/2dust/v2rayNG)
- [HarmonyOS VPN API](https://developer.harmonyos.com/)
