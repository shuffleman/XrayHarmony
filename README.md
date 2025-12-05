# XrayHarmony

<p align="center">
  <strong>Xray-core 的鸿蒙系统封装</strong>
</p>

<p align="center">
  为 HarmonyOS 应用提供完整的 Xray-core 代理功能
</p>

## 📖 简介

XrayHarmony 是一个为鸿蒙系统（HarmonyOS）设计的 [Xray-core](https://github.com/xtls/xray-core) 封装库。它提供了从底层 Go 实现到高层 ArkTS 接口的完整封装，使得鸿蒙应用开发者可以轻松集成 Xray 的强大代理功能。

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

- Go 1.21 或更高版本
- HarmonyOS SDK
- CMake 3.16 或更高版本
- GCC/Clang 编译器
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

### VPN 模式 (推荐)

XrayHarmony 现已支持完整的 VPN 功能，可实现系统级全局代理：

```typescript
import VpnExtensionAbility from '@ohos.app.ability.VpnExtensionAbility';
import vpnExt from '@ohos.net.vpnExtension';
import { XrayClient, createXrayClient } from './index';
import { XrayVPNClient, createXrayVPNClient, VPNConfig } from './vpn';

export default class XrayVpnExtension extends VpnExtensionAbility {
  private xrayClient: XrayClient;
  private vpnClient: XrayVPNClient;
  private vpnConnection: vpnExt.VpnConnection;

  async startVPN(xrayConfig: any): Promise<void> {
    // 1. 创建并启动 Xray
    this.xrayClient = createXrayClient();
    await this.xrayClient.loadConfig(xrayConfig);
    await this.xrayClient.start();

    // 2. 创建 TUN 设备
    this.vpnConnection = vpnExt.createVpnConnection(this.context);
    const tunConfig = {
      addresses: [{ address: { address: '10.0.0.2', family: 1 }, prefixLength: 24 }],
      routes: [{ interface: 'vpn-tun', destination: { address: '0.0.0.0', family: 1 }, prefixLength: 0 }],
      mtu: 1400,
      dnsAddresses: [{ address: '8.8.8.8', family: 1 }]
    };
    const tunFd = await this.vpnConnection.create(tunConfig);

    // 3. 启动 VPN
    this.vpnClient = createXrayVPNClient(this.xrayClient.instanceId);
    await this.vpnClient.start({
      tunFd: tunFd,
      tunMTU: 1400,
      socksAddr: '127.0.0.1:10808',
      dnsServers: ['8.8.8.8', '8.8.4.4']
    });
  }
}
```

详细的 VPN 使用指南请参考 [VPN 文档](docs/VPN.md)。

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

- [API 文档](docs/API.md) - 完整的 API 参考
- [VPN 使用指南](docs/VPN.md) - TUN + Xray VPN 功能详细说明
- [构建文档](docs/BUILD.md) - 构建和集成指南
- [示例代码](examples/) - 各种使用场景示例

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

## 🙏 致谢

- [Xray-core](https://github.com/xtls/xray-core) - 强大的代理工具核心
- [tun2socks](https://github.com/xjasonlyu/tun2socks) - 优秀的 TUN 网络栈实现
- HarmonyOS 开发团队 - 提供优秀的开发平台

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
