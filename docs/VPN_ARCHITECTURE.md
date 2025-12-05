# XrayHarmony VPN 架构说明

## 架构设计

XrayHarmony VPN 功能采用**分离式架构**，避免依赖冲突：

```
┌────────────────────────────────────────────┐
│         HarmonyOS 应用层                    │
│  ┌──────────────────────────────────────┐  │
│  │  VpnExtensionAbility                 │  │
│  │  - 创建 TUN 设备                      │  │
│  │  - 获取 TUN 文件描述符 (fd)          │  │
│  └────────┬─────────────────────────────┘  │
└───────────┼────────────────────────────────┘
            │
            ├──> TUN fd
            │
┌───────────┼────────────────────────────────┐
│  ┌────────▼───────────┐  ┌──────────────┐ │
│  │   tun2socks        │  │  Xray-core   │ │
│  │   (独立进程/库)     │──>│  SOCKS5入站  │ │
│  │                    │  │              │ │
│  │  - 接收 TUN fd     │  │  - 代理出站  │ │
│  │  - SOCKS5 客户端   │  │              │ │
│  └────────────────────┘  └──────────────┘ │
│                                            │
│         Native 层 (分离部署)                │
└────────────────────────────────────────────┘
            │
            ↓
       网络流量代理
```

## 为什么采用分离式架构？

### 依赖冲突问题

1. **tun2socks v2.6.0** 依赖 **gvisor v0.0.0-20250523182742** (新版本)
2. **Xray-core v1.8.16** 依赖 **gvisor v0.0.0-20231202080848** (旧版本)
3. 在同一个 Go 模块中无法同时满足两者的依赖要求

### 解决方案

将 tun2socks 和 Xray-core 分离为**独立的组件**：

- **XrayHarmony**: 只封装 Xray-core，提供 SOCKS5 入站
- **tun2socks**: 作为独立的二进制或库，处理 TUN 流量转发

## 实现方案

### 方案 A：使用独立的 tun2socks 二进制（推荐）

1. **编译 tun2socks**
   ```bash
   git clone https://github.com/xjasonlyu/tun2socks
   cd tun2socks
   GOOS=linux GOARCH=arm64 go build -o tun2socks
   ```

2. **在 HarmonyOS 应用中启动 tun2socks**
   ```typescript
   // 1. 启动 Xray (SOCKS5 入站在 127.0.0.1:10808)
   const xrayClient = createXrayClient();
   await xrayClient.loadConfig(xrayConfig);
   await xrayClient.start();

   // 2. 创建 TUN 设备
   const vpnConnection = vpnExt.createVpnConnection(this.context);
   const tunFd = await vpnConnection.create(tunConfig);

   // 3. 启动 tun2socks (通过 Bash 或 Native 调用)
   // tun2socks -device fd://<tunFd> -proxy socks5://127.0.0.1:10808 -mtu 1400
   ```

### 方案 B：使用预编译的 tun2socks 库

1. 从 [tun2socks releases](https://github.com/xjasonlyu/tun2socks/releases) 下载预编译的库
2. 集成到 HarmonyOS 项目
3. 通过 FFI 调用

### 方案 C：自己实现简单的 TUN -> SOCKS5 转发

使用 Go 或 C++ 实现一个简单的数据包转发器，避免使用 tun2socks。

## XrayHarmony 的职责

XrayHarmony 专注于：

1. **封装 Xray-core**：提供完整的 Xray 代理功能
2. **SOCKS5 入站**：监听本地端口（如 10808）
3. **代理管理**：配置、启动、停止、统计

**不包含**：
- TUN 设备管理（由 HarmonyOS VpnExtensionAbility 处理）
- TUN 流量转发（由独立的 tun2socks 处理）

## 配置示例

### Xray 配置（必须包含 SOCKS5 入站）

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
      "settings": { /* 你的代理配置 */ }
    }
  ]
}
```

### tun2socks 启动参数

```bash
tun2socks \
  -device fd://<tunFd> \
  -proxy socks5://127.0.0.1:10808 \
  -mtu 1400 \
  -loglevel info
```

## 完整流程

```
1. HarmonyOS VpnExtensionAbility 创建 TUN 设备 → 获取 tunFd
                                                       ↓
2. 启动 Xray-core (SOCKS5 入站在 127.0.0.1:10808)  ←─┐
                                                       │
3. 启动 tun2socks (fd://<tunFd> → socks5://127.0.0.1:10808)
                                                       │
4. TUN 流量 → tun2socks → Xray SOCKS5 → 代理出站 ────┘
```

## 优势

1. **无依赖冲突**：tun2socks 和 Xray-core 分别编译
2. **模块化**：每个组件职责清晰
3. **灵活性**：可以独立升级 tun2socks 或 Xray-core
4. **稳定性**：避免 gvisor 版本冲突导致的编译错误

## 参考资源

- [tun2socks GitHub](https://github.com/xjasonlyu/tun2socks)
- [Xray-core 文档](https://xtls.github.io/)
- [HarmonyOS VPN API](https://developer.harmonyos.com/)
