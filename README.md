# XrayHarmony

<p align="center">
  <strong>Xray-core 的鸿蒙系统封装</strong>
</p>

<p align="center">
  为 HarmonyOS 应用提供完整的 Xray-core 代理功能
</p>

## 📖 简介

XrayHarmony 是一个为鸿蒙系统（HarmonyOS）设计的 [Xray-core](https://github.com/xtls/xray-core) 封装库。它提供了从底层 Go 实现到高层 ArkTS 接口的完整封装，使得鸿蒙应用开发者可以轻松集成 Xray 的强大代理功能。

### 🎉 项目状态

- **当前版本**: 基于 Xray-core v1.251202.0
- **Go 版本**: 1.25 (toolchain go1.25.5)
- **核心状态**: ✅ Xray-core 封装已完成并稳定运行
- **VPN 功能**: ✅ 完整支持 TUN 网卡和系统级 VPN
- **依赖状态**: ✅ 所有依赖冲突已解决

## ✨ 特性

- 🎯 **完整封装**：从 Go 到 ArkTS 的多层封装架构
- 🚀 **易于使用**：简洁的 ArkTS API，符合鸿蒙开发习惯
- 🔒 **类型安全**：完整的 TypeScript 类型定义
- 📱 **原生性能**：基于 C/C++ 桥接的高性能实现
- 🛠️ **灵活配置**：支持 JSON 配置和文件配置
- 📊 **实时统计**：提供运行状态和流量统计
- 🎨 **多架构支持**：支持 ARM64、ARM、AMD64 等多种架构
- 🌐 **VPN 模式**：支持 TUN 网卡，实现完整的系统级 VPN 功能

## 🏗️ 架构

```
┌─────────────────────────────────┐
│   HarmonyOS Application (ArkTS)  │  ← 应用层
├─────────────────────────────────┤
│   XrayHarmony ArkTS Interface   │  ← TypeScript 接口层
├─────────────────────────────────┤
│   Native C++ Bridge Layer       │  ← C++ 桥接层
├─────────────────────────────────┤
│   Go Wrapper (CGO)              │  ← Go 封装层
├─────────────────────────────────┤
│   Xray-core                     │  ← Xray 核心
└─────────────────────────────────┘
```

## 📦 项目结构

```
XrayHarmony/
├── go/                     # Go 封装层
│   ├── wrapper/           # Xray-core 封装
│   │   ├── xray_wrapper.go
│   │   └── export.go      # C 导出接口
│   └── go.mod
├── native/                # C++ 桥接层
│   ├── include/          # 头文件
│   │   └── xray_bridge.h
│   └── src/              # 实现文件
│       └── xray_bridge.cpp
├── arkts/                # ArkTS 接口层
│   ├── src/
│   │   ├── index.ets     # 主接口
│   │   └── index.d.ts    # 类型定义
│   └── package.json
├── build/                # 构建脚本
│   ├── build.sh          # Go 库构建脚本
│   └── CMakeLists.txt    # CMake 配置
├── examples/             # 示例代码
│   ├── basic_usage.ets
│   └── config.json
├── docs/                 # 文档
│   └── API.md           # API 文档
├── Makefile             # Make 构建配置
└── README.md
```

## 🚀 快速开始

### 前置要求

- **Go 1.25 或更高版本** (推荐 1.25.5)
- HarmonyOS SDK
- CMake 3.16 或更高版本
- GCC/Clang 编译器（支持 C++17）
- 交叉编译工具链（用于目标架构）

### 编译

1. **克隆仓库**

```bash
git clone https://github.com/shuffleman/XrayHarmony.git
cd XrayHarmony
```

2. **安装依赖**

```bash
make install
```

3. **构建所有平台**

```bash
make all
```

或构建特定架构：

```bash
# ARM64
make build-go ARCH=arm64

# AMD64
make build-go ARCH=amd64

# ARM
make build-go ARCH=arm
```

4. **构建结果**

编译完成后，共享库将位于 `libs/` 目录：

```
libs/
├── libxray_linux_arm64.so
├── libxray_linux_amd64.so
└── libxray_linux_arm.so
```

### 集成到 HarmonyOS 项目

1. **复制库文件**

将编译好的 `.so` 文件复制到你的 HarmonyOS 项目的 `libs/` 目录。

2. **复制 ArkTS 接口**

将 `arkts/src/` 目录下的文件复制到你的项目中。

3. **在代码中使用**

```typescript
import { XrayClient, XrayConfig } from './path/to/index.ets';

// 创建客户端
const client = new XrayClient();

// 配置
const config: XrayConfig = {
  inbound: {
    protocol: 'socks',
    port: 1080,
    listen: '127.0.0.1'
  },
  outbound: {
    protocol: 'freedom'
  },
  log: {
    loglevel: 'info'
  }
};

// 加载配置并启动
await client.loadConfig(config);
await client.start();

// 检查状态
if (client.isRunning()) {
  console.log('Xray is running!');
}

// 获取统计
const stats = await client.getStats();
console.log('Stats:', stats);

// 停止并清理
await client.stop();
client.destroy();
```

## 📚 使用示例

### 快速开始 - 解析分享链接

```typescript
import { createXrayClient } from '@shuffleman/xray-harmony';

const client = createXrayClient();

// 解析 VMess/VLESS/Trojan/SS 分享链接
const shareURL = "vmess://eyJ2IjoiMiIsInBzIjoi...";
const serverConfig = await client.parseShareURL(shareURL);

console.log('服务器:', serverConfig.address);
console.log('端口:', serverConfig.port);
console.log('协议:', serverConfig.protocol);
```

### 使用 JSON 配置

```typescript
// 直接使用 Xray 标准 JSON 配置
const configJSON = JSON.stringify({
  log: { loglevel: 'warning' },
  inbounds: [{
    protocol: 'socks',
    listen: '127.0.0.1',
    port: 10808,
    settings: { auth: 'noauth', udp: true }
  }],
  outbounds: [{
    protocol: 'vmess',
    settings: {
      vnext: [{
        address: 'server.example.com',
        port: 443,
        users: [{ id: 'your-uuid-here', alterId: 0, security: 'auto' }]
      }]
    }
  }, {
    tag: 'direct',
    protocol: 'freedom'
  }],
  routing: {
    rules: [{
      type: 'field',
      outboundTag: 'direct',
      domain: ['geosite:cn'],
      ip: ['geoip:cn', 'geoip:private']
    }]
  },
  stats: {}
});

await client.loadConfig(configJSON);
await client.start();
```

### 资产管理

```typescript
import { AssetManager } from '@shuffleman/xray-harmony';

const assetMgr = new AssetManager('/data/storage/el2/base/assets');

// 下载 geoip 和 geosite
await assetMgr.download('geoip', '', (progress) => {
  console.log(`下载进度: ${progress.percentage}%`);
});

await assetMgr.download('geosite', '');

// geoip/geosite 文件会被自动使用在路由规则中
```

### VPN 模式

XrayHarmony 内置 tun2socks 封装,可以实现系统级 VPN 功能。

**架构说明**：
```
HarmonyOS VPN API → TUN 设备
         ↓
    Tun2Socks (内置) → SOCKS5 连接
         ↓
    Xray (SOCKS5 入站) → 代理服务器
```

**完整实现**：

```typescript
import { createXrayClient, Tun2Socks } from '@shuffleman/xray-harmony';
import vpnExtension from '@ohos.net.vpnExtension';

// 1. 启动 Xray SOCKS5 代理
const client = createXrayClient();
await client.loadConfig({
  inbound: {
    protocol: 'socks',
    port: 10808,
    listen: '127.0.0.1',
    settings: { auth: 'noauth', udp: true }
  },
  outbound: {
    protocol: 'vmess',
    settings: { /* 你的服务器配置 */ }
  }
});
await client.start();

// 2. 创建 VPN 连接并获取 TUN FD
const vpnConnection = vpnExtension.createVpnConnection(getContext());
await vpnConnection.setUp({
  addresses: [{ address: { address: '10.0.0.2' }, prefixLength: 24 }],
  routes: [{ interface: 'tun0', destination: { address: '0.0.0.0' }, prefixLength: 0 }],
  dnsServers: ['8.8.8.8'],
  mtu: 1500
});
const tunFd = vpnConnection.getFileDescriptor();

// 3. 启动 Tun2Socks
const tun2socks = new Tun2Socks({
  tunFd: tunFd,
  socksAddr: '127.0.0.1:10808',
  mtu: 1500,
  dnsAddr: '8.8.8.8:53'
});
await tun2socks.start();

console.log('VPN 已启动!');

// 4. 获取统计信息
const stats = await client.getStats();
const tunStats = await tun2socks.getStats();
console.log('Xray 统计:', stats);
console.log('隧道统计:', tunStats);
```

详细的 VPN 实现指南请参考：
- [完整 API 文档](docs/XRAY_WRAPPER_API.md) - 所有功能的详细说明
- [集成指南](docs/INTEGRATION_GUIDE_CN.md) - 完整的集成步骤
- [VPN 架构文档](docs/VPN_ARCHITECTURE.md) - 架构设计和实现方案
- [VPNControl_Demo](examples/VPNControl_Demo/) - 完整示例项目

### 基础使用

```typescript
import { createXrayClient, XrayConfig } from '@shuffleman/xray-harmony';

const client = createXrayClient();

const config: XrayConfig = {
  inbound: {
    protocol: 'socks',
    port: 1080,
    listen: '127.0.0.1',
    settings: {
      auth: 'noauth',
      udp: true
    }
  },
  outbound: {
    protocol: 'freedom',
    settings: {}
  }
};

try {
  await client.loadConfig(config);
  await client.start();
  console.log('Xray started successfully');
} catch (error) {
  console.error('Error:', error);
} finally {
  client.destroy();
}
```

### 从文件加载配置

```typescript
const client = createXrayClient();

try {
  await client.loadConfigFromFile('/data/storage/el2/base/xray_config.json');
  await client.start();
} catch (error) {
  console.error('Error:', error);
}
```

### 服务封装

```typescript
import { XrayClient, XrayConfig } from '@shuffleman/xray-harmony';

export class XrayService {
  private client: XrayClient;

  async initialize(config: XrayConfig): Promise<boolean> {
    this.client = new XrayClient();
    try {
      await this.client.loadConfig(config);
      return true;
    } catch {
      return false;
    }
  }

  async start(): Promise<boolean> {
    try {
      await this.client.start();
      return true;
    } catch {
      return false;
    }
  }

  async stop(): Promise<boolean> {
    try {
      await this.client.stop();
      return true;
    } catch {
      return false;
    }
  }

  isRunning(): boolean {
    return this.client?.isRunning() ?? false;
  }
}
```

更多示例请查看 [examples/](examples/) 目录。

## 📖 文档

### 核心文档
- [完整 API 文档](docs/XRAY_WRAPPER_API.md) - **新!** 所有功能的详细 API 说明
- [集成指南](docs/INTEGRATION_GUIDE_CN.md) - **新!** 完整的集成步骤和示例
- [API 参考](docs/API.md) - 基础 API 参考
- [构建文档](docs/BUILD.md) - 构建和编译指南

### VPN 相关
- [VPN 架构文档](docs/VPN_ARCHITECTURE.md) - VPN 技术架构说明
- [VPN 使用指南](docs/VPN.md) - TUN + Xray VPN 功能详细说明
- [VPN 示例项目](examples/VPNControl_Demo/) - 完整的鸿蒙 VPN 示例应用

### 其他
- [升级记录](UPGRADE_PLAN.md) - Xray-core 升级历史和当前版本信息
- [示例代码](examples/) - 各种使用场景示例

### 新功能特性

#### 🔧 协议工具
支持解析和生成主流代理协议的分享链接:
- VMess (v2rayN 格式)
- VLESS (标准格式)
- Trojan (标准格式)
- Shadowsocks (标准格式)

#### 🌐 Tun2Socks
内置 tun2socks 封装,无需外部依赖:
- 处理 TUN 设备流量
- 转发到 SOCKS5 代理
- 实时流量统计
- 支持 UDP

#### 📦 资产管理
自动管理路由规则数据:
- geoip.dat (IP 数据库)
- geosite.dat (域名数据库)
- 自动检查更新
- 下载进度跟踪

#### ⚙️ 标准配置
使用 Xray 标准 JSON 配置:
- 完全兼容 Xray 官方配置
- 支持所有协议和功能
- 灵活的配置方式
- 易于从其他项目迁移

## 🔧 开发

### 构建命令

```bash
# 显示帮助
make help

# 构建所有
make all

# 只构建 Go 库
make build-go

# 只构建 Native 层
make build-native

# 清理
make clean

# 运行测试
make test
```

### 目录说明

- `go/` - Go 语言的 Xray-core 封装层
- `native/` - C++ 桥接层，连接 Go 和 ArkTS
- `arkts/` - ArkTS/TypeScript 接口层
- `build/` - 构建脚本和配置
- `examples/` - 使用示例
- `docs/` - 文档

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)（待添加）了解详情。

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## ⚠️ 免责声明

本项目仅供学习和研究使用。使用本项目时，请遵守当地法律法规。开发者不对使用本项目造成的任何后果负责。

## 🚀 技术栈

### 核心技术
- **Xray-core**: v1.251202.0 - 最新稳定版本
- **Go**: 1.25 with toolchain go1.25.5
- **gvisor**: v0.0.0-20250428193742-2d800c3129d5 - 网络栈支持
- **HarmonyOS**: 支持 API 9+

### 架构特点
- **零依赖冲突**: 所有依赖版本兼容性已验证
- **原生性能**: 基于 CGO 的高效 C/C++ 桥接
- **模块化设计**: Go → C++ → ArkTS 清晰的分层架构
- **完整代理支持**: 支持所有 Xray 协议（SOCKS5、VMess、VLESS、Trojan 等）

### 🎯 v2.0.0 新增功能

#### 1. **协议工具** (参考 v2rayNG 实现)
- ✅ VMess 链接解析和生成 (`vmess://`)
- ✅ VLESS 链接解析和生成 (`vless://`)
- ✅ Trojan 链接解析和生成 (`trojan://`)
- ✅ Shadowsocks 链接解析和生成 (`ss://`)
- ✅ 自动识别协议类型

#### 2. **Tun2Socks 封装**
- ✅ 完整的 tun2socks 框架
- ✅ TUN 设备流量处理
- ✅ SOCKS5 代理转发
- ✅ 流量统计功能
- ✅ 支持 VPN 模式

#### 3. **资产管理器**
- ✅ geoip.dat 管理和下载
- ✅ geosite.dat 管理和下载
- ✅ 自动检查更新
- ✅ 文件验证
- ✅ 下载进度回调

#### 4. **标准配置**
- ✅ 使用 Xray 标准 JSON 配置
- ✅ 完全兼容 Xray 官方配置
- ✅ 支持所有协议和功能
- ✅ 灵活的配置方式

#### 5. **增强的统计功能**
- ✅ 实时流量统计
- ✅ 运行时长记录
- ✅ 上行/下行字节数
- ✅ 分组件统计 (Xray + Tun2Socks)

### VPN 功能说明
- **核心封装**: XrayHarmony 封装 Xray-core 提供 SOCKS5 代理功能
- **Tun2Socks**: 内置 tun2socks 封装,处理 TUN 设备流量
- **完整方案**: HarmonyOS VPN API → TUN → Tun2Socks → Xray → 远程服务器
- **参考示例**: 查看 `examples/VPNControl_Demo` 了解 VPN 集成方案

## 🙏 致谢

- [Xray-core](https://github.com/xtls/xray-core) - 强大的代理工具核心
- [tun2socks](https://github.com/xjasonlyu/tun2socks) - TUN 流量处理
- [gvisor](https://github.com/google/gvisor) - 高性能用户态网络栈
- HarmonyOS 开发团队 - 提供优秀的开发平台
- 所有贡献者和使用者

## 📮 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/shuffleman/XrayHarmony/issues)
- 发起 [Pull Request](https://github.com/shuffleman/XrayHarmony/pulls)

## 🌟 Star History

如果这个项目对你有帮助，请给它一个 Star ⭐️

---

<p align="center">
  Made with ❤️ for HarmonyOS
</p>
