# XrayHarmony VPN 使用指南

本文档介绍如何在 HarmonyOS 上使用 XrayHarmony 实现完整的 VPN 功能。

## 目录

- [架构概述](#架构概述)
- [前置要求](#前置要求)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API 参考](#api-参考)
- [常见问题](#常见问题)

## 架构概述

XrayHarmony VPN 功能基于以下架构：

```
┌─────────────────────────────────────────────────┐
│         HarmonyOS 应用层                         │
│  ┌──────────────────────────────────────────┐   │
│  │    VpnExtensionAbility (系统 VPN API)    │   │
│  └─────────────┬────────────────────────────┘   │
│                │ 获取 TUN 文件描述符              │
│                ↓                                 │
│  ┌──────────────────────────────────────────┐   │
│  │     XrayVPNClient (ArkTS 封装层)         │   │
│  └─────────────┬────────────────────────────┘   │
└────────────────┼──────────────────────────────────┘
                 │ Native 调用
┌────────────────┼──────────────────────────────────┐
│                ↓                                  │
│  ┌──────────────────────────────────────────┐   │
│  │      VPNBridge (C++ 桥接层)              │   │
│  └─────────────┬────────────────────────────┘   │
│                │                                  │
│                ↓                                  │
│  ┌──────────────────────────────────────────┐   │
│  │    VPNManager (Go 管理层)                │   │
│  │  ┌────────────┐      ┌────────────────┐  │   │
│  │  │ tun2socks  │──────→│ Xray SOCKS5   │  │   │
│  │  │  (网络栈)   │      │   (代理核心)   │  │   │
│  │  └────────────┘      └────────────────┘  │   │
│  └──────────────────────────────────────────┘   │
│                                                  │
│           libxray.so (共享库)                    │
└──────────────────────────────────────────────────┘
                 ↓
           网络流量代理
```

## 前置要求

### 1. 权限声明

在 `module.json5` 中声明必要的权限：

```json
{
  "module": {
    "requestPermissions": [
      {
        "name": "ohos.permission.INTERNET",
        "reason": "需要访问网络"
      },
      {
        "name": "ohos.permission.GET_NETWORK_INFO",
        "reason": "需要获取网络信息"
      }
    ]
  }
}
```

**注意**: VPN 相关权限目前仅对系统应用开放，需要通过 ACL 申请。

### 2. VpnExtensionAbility 配置

在 `module.json5` 中配置 VPN 扩展能力：

```json
{
  "module": {
    "extensionAbilities": [
      {
        "name": "XrayVpnExtension",
        "srcEntry": "./ets/vpnextension/XrayVpnExtension.ets",
        "type": "vpn.extension",
        "exported": true
      }
    ]
  }
}
```

### 3. 依赖库

确保项目包含 XrayHarmony 的所有依赖：

```
libs/
├── libxray_linux_arm64.so
└── libxray_linux_arm64.h
```

## 快速开始

### 1. 启动 Xray SOCKS5 代理

首先启动 XrayHarmony 提供 SOCKS5 入站：

```typescript
import { createXrayClient } from '../arkts/src/index';

export default class XrayVpnExtension extends VpnExtensionAbility {
  private xrayClient: XrayClient | null = null;
  private vpnConnection: vpnExt.VpnConnection | null = null;
  private tun2socksProcess: any = null;  // tun2socks 进程引用

  onCreate(): void {
    // 创建 VPN 连接对象
    this.vpnConnection = vpnExt.createVpnConnection(this.context);
  }

  async startVPN(xrayConfig: any): Promise<void> {
    // 1. 创建并启动 Xray (SOCKS5 入站)
    this.xrayClient = createXrayClient();
    await this.xrayClient.loadConfig(xrayConfig);
    await this.xrayClient.start();

    // 2. 创建 TUN 设备
    const tunConfig: vpnExt.VpnConfig = {
      addresses: [{ address: { address: '10.0.0.2', family: 1 }, prefixLength: 24 }],
      routes: [{ interface: 'vpn-tun', destination: { address: '0.0.0.0', family: 1 }, prefixLength: 0 }],
      mtu: 1400,
      dnsAddresses: [{ address: '8.8.8.8', family: 1 }]
    };
    const tunFd = await this.vpnConnection!.create(tunConfig);

    // 3. 启动 tun2socks（需要单独实现或集成）
    // 这里只是示意，实际需要根据项目情况实现 tun2socks 的启动
    // 方式1: 作为独立进程启动
    // 方式2: 作为 Native 库集成
    // 详见 VPN_ARCHITECTURE.md 文档
    await this.startTun2Socks(tunFd, '127.0.0.1:10808', 1400);
  }

  async startTun2Socks(tunFd: number, socksAddr: string, mtu: number): Promise<void> {
    // TODO: 实现 tun2socks 启动逻辑
    // 可以使用独立的 tun2socks 二进制或库
    // 参考 examples/VPNControl_Demo 中的实现
  }
}
```

### 2. 准备 Xray 配置

创建 Xray 配置文件（参考 `examples/vpn_config.json`）：

```json
{
  "inbounds": [
    {
      "tag": "socks-in",
      "port": 10808,
      "listen": "127.0.0.1",
      "protocol": "socks",
      "settings": {
        "auth": "noauth",
        "udp": true
      },
      "sniffing": {
        "enabled": true,
        "destOverride": ["http", "tls"]
      }
    }
  ],
  "outbounds": [
    {
      "tag": "proxy",
      "protocol": "vmess",
      "settings": { /* 你的代理服务器配置 */ }
    }
  ]
}
```

**关键点**：
- `inbounds` 必须包含一个 SOCKS5 入站（默认端口 10808）
- `sniffing` 启用以支持流量嗅探和域名解析
- `udp` 设置为 true 以支持 UDP 流量

### 3. 集成 tun2socks

有几种方式集成 tun2socks：

**方式 A：独立二进制进程**
```bash
# 编译 tun2socks
git clone https://github.com/xjasonlyu/tun2socks
cd tun2socks
GOOS=linux GOARCH=arm64 go build -o tun2socks

# 在应用中启动
tun2socks -device fd://<tunFd> -proxy socks5://127.0.0.1:10808 -mtu 1400
```

**方式 B：使用预编译库**

从 [tun2socks releases](https://github.com/xjasonlyu/tun2socks/releases) 下载预编译库并集成。

### 4. 启动 VPN

完整流程：

```typescript
// 1. 启动 Xray SOCKS5
const xrayClient = createXrayClient();
await xrayClient.loadConfig(xrayConfig);
await xrayClient.start();

// 2. 创建 TUN 设备（HarmonyOS VPN API）
const tunFd = await vpnConnection.create(tunConfig);

// 3. 启动 tun2socks
await startTun2Socks(tunFd, '127.0.0.1:10808', 1400);

// VPN 现在已启动并运行
```

## 配置说明

### Xray 配置

Xray 配置遵循标准的 Xray-core 配置格式，必须包含：

#### 1. SOCKS5 入站

```json
{
  "tag": "socks-in",
  "port": 10808,
  "listen": "127.0.0.1",
  "protocol": "socks",
  "settings": {
    "auth": "noauth",
    "udp": true,
    "ip": "127.0.0.1"
  },
  "sniffing": {
    "enabled": true,
    "destOverride": ["http", "tls"]
  }
}
```

#### 2. 出站代理

支持所有 Xray 协议：
- VMess
- VLESS
- Trojan
- Shadowsocks
- Freedom (直连)
- Blackhole (阻止)

#### 3. 路由规则

```json
{
  "routing": {
    "domainStrategy": "IPIfNonMatch",
    "rules": [
      { "type": "field", "ip": ["geoip:private"], "outboundTag": "direct" },
      { "type": "field", "ip": ["geoip:cn"], "outboundTag": "direct" },
      { "type": "field", "domain": ["geosite:cn"], "outboundTag": "direct" }
    ]
  }
}
```

### tun2socks 配置参数

tun2socks 的主要配置参数：

```bash
tun2socks \
  -device fd://<tunFd> \      # TUN 设备文件描述符
  -proxy socks5://127.0.0.1:10808 \  # SOCKS5 代理地址
  -mtu 1400 \                 # MTU 值
  -loglevel info \            # 日志级别
  -stats :8080                # 可选：统计信息端口
```

### TUN 设备配置（HarmonyOS VPN API）

```typescript
interface VpnConfig {
  addresses: VpnAddress[];     // VPN 虚拟网卡 IP 地址
  routes: VpnRoute[];          // 路由配置
  mtu: number;                 // MTU 值
  dnsAddresses: VpnAddress[];  // DNS 服务器
  trustedApplications?: string[];  // 应用白名单
  blockedApplications?: string[];  // 应用黑名单
}
```

**推荐配置**：
- IP 地址: 10.0.0.2/24
- 网关: 10.0.0.1
- MTU: 1400
- 默认路由: 0.0.0.0/0（全局代理）
- DNS: 8.8.8.8, 8.8.4.4

## API 参考

### XrayClient (基础代理功能)

XrayHarmony 提供的核心 API，用于管理 Xray 代理：

```typescript
import { createXrayClient, XrayClient } from '@shuffleman/xray-harmony';

const client = createXrayClient();

// 加载配置
await client.loadConfig(config);

// 启动代理
await client.start();

// 检查状态
const isRunning = client.isRunning();

// 停止代理
await client.stop();

// 释放资源
client.destroy();
```

详细的 API 说明请参考 [API 文档](API.md)。

### tun2socks 集成

tun2socks 需要单独集成，XrayHarmony 不提供封装的 API。集成方式：

1. **进程方式**：使用 HarmonyOS 进程 API 启动 tun2socks 二进制
2. **库方式**：通过 Native 层集成 tun2socks 库
3. **参考实现**：查看 `examples/VPNControl_Demo` 示例

## 常见问题

### 1. 如何调试 VPN 连接？

启用详细日志：

```json
{
  "log": {
    "loglevel": "debug"
  }
}
```

查看日志：
```typescript
console.info('[XrayVPN] VPN status:', await vpnClient.getStats());
```

### 2. VPN 无法连接？

检查以下几点：
1. 确认 Xray 配置正确，特别是 SOCKS5 入站端口
2. 确认代理服务器配置正确
3. 检查 TUN 设备是否创建成功
4. 查看错误日志：`vpnClient.getLastError()`

### 3. 部分应用无法代理？

使用应用白名单/黑名单：

```typescript
const tunConfig: vpnExt.VpnConfig = {
  // ...其他配置
  trustedApplications: ['com.example.app1', 'com.example.app2'],
  // 或者
  blockedApplications: ['com.example.excluded']
};
```

### 4. UDP 流量无法代理？

确保：
1. Xray 配置中启用了 UDP：`"udp": true`
2. VPN 配置中启用了 UDP：`udp: true`
3. 代理服务器支持 UDP

### 5. 性能优化建议

1. **MTU 调优**: 根据网络环境调整 MTU 值（通常 1400-1500）
2. **TCP 并发**: 对于高延迟网络，启用 `tcpConcurrent`
3. **路由优化**: 使用精确的路由规则，避免不必要的代理流量

### 6. 如何实现分应用代理？

使用 HarmonyOS 的应用过滤功能：

```typescript
const tunConfig: vpnExt.VpnConfig = {
  addresses: [...],
  routes: [...],
  mtu: 1400,
  dnsAddresses: [...],

  // 只代理这些应用
  trustedApplications: [
    'com.example.browser',
    'com.example.social'
  ]
};
```

### 7. 如何实现国内外分流？

在 Xray 配置中使用路由规则：

```json
{
  "routing": {
    "domainStrategy": "IPIfNonMatch",
    "rules": [
      {
        "type": "field",
        "ip": ["geoip:cn", "geoip:private"],
        "outboundTag": "direct"
      },
      {
        "type": "field",
        "domain": ["geosite:cn"],
        "outboundTag": "direct"
      },
      {
        "type": "field",
        "network": "tcp,udp",
        "outboundTag": "proxy"
      }
    ]
  }
}
```

## 参考资源

- [Xray-core 官方文档](https://xtls.github.io/)
- [HarmonyOS VPN API 文档](https://developer.harmonyos.com/)
- [tun2socks 项目](https://github.com/xjasonlyu/tun2socks)

## 许可证

本项目遵循与主项目相同的许可证。
