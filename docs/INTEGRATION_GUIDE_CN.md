# XrayHarmony 集成指南

本指南详细说明如何将 XrayHarmony 集成到 HarmonyOS 项目中,以及如何使用完整的 Xray 功能。

## 目录

- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [编译构建](#编译构建)
- [集成步骤](#集成步骤)
- [功能使用](#功能使用)
- [VPN 实现](#vpn-实现)
- [常见问题](#常见问题)

## 快速开始

### 前置要求

- Go 1.25+
- HarmonyOS SDK
- NDK (用于交叉编译)
- CMake 3.16+

### 克隆项目

```bash
git clone https://github.com/shuffleman/XrayHarmony.git
cd XrayHarmony
```

### 编译

```bash
# 安装 Go 依赖
make install

# 编译所有平台 (arm64, amd64, arm)
make all

# 或编译特定平台
make build-go ARCH=arm64
```

编译完成后,库文件位于 `libs/` 目录:
- `libxray_linux_arm64.so`
- `libxray_linux_amd64.so`
- `libxray_linux_arm.so`

## 项目结构

```
XrayHarmony/
├── go/                          # Go 封装层
│   └── wrapper/
│       ├── xray_wrapper.go      # Xray 核心封装
│       ├── export.go            # C 导出接口
│       ├── config_builder.go    # 配置构建器
│       ├── protocol_utils.go    # 协议解析工具
│       ├── tun2socks.go         # Tun2socks 封装
│       └── asset_manager.go     # 资产管理
├── native/                      # C++ 桥接层
│   ├── include/xray_bridge.h    # 头文件
│   └── src/xray_bridge.cpp      # 实现
├── arkts/                       # ArkTS 接口层
│   └── src/
│       ├── index.ets            # 主接口
│       └── index.d.ts           # 类型定义
├── libs/                        # 编译产物
└── docs/                        # 文档
    ├── XRAY_WRAPPER_API.md      # 完整 API 文档
    └── INTEGRATION_GUIDE_CN.md  # 本指南
```

## 编译构建

### Go 层编译

```bash
# 编译为共享库
cd go
CGO_ENABLED=1 GOOS=linux GOARCH=arm64 \
  go build -buildmode=c-shared -o ../libs/libxray_linux_arm64.so wrapper/*.go
```

### 主要组件

项目包含以下主要组件:

1. **Xray 核心封装** (`xray_wrapper.go`): Xray-core 生命周期管理
2. **配置构建器** (`config_builder.go`): 流畅的配置 API
3. **协议工具** (`protocol_utils.go`): 解析和生成分享链接
4. **Tun2Socks** (`tun2socks.go`): VPN 流量处理
5. **资产管理** (`asset_manager.go`): geoip/geosite 管理

## 集成步骤

### 1. 复制库文件

将编译好的 `.so` 文件复制到 HarmonyOS 项目:

```
YourHarmonyProject/
└── entry/
    └── libs/
        └── arm64-v8a/
            └── libxray.so
```

### 2. 配置 build-profile.json5

```json5
{
  "buildOption": {
    "externalNativeOptions": {
      "path": "./src/main/cpp/CMakeLists.txt",
      "arguments": "",
      "cppFlags": "",
    }
  }
}
```

### 3. 配置 CMakeLists.txt

```cmake
cmake_minimum_required(VERSION 3.16)
project(YourProject)

# 添加 libxray
add_library(xray SHARED IMPORTED)
set_target_properties(xray PROPERTIES
    IMPORTED_LOCATION ${CMAKE_CURRENT_SOURCE_DIR}/../../../libs/${OHOS_ARCH}/libxray.so
)

# 链接
target_link_libraries(entry PUBLIC xray)
```

### 4. 导入 ArkTS 接口

将 `arkts/src/` 下的文件复制到项目中,或作为模块引入:

```typescript
import { XrayClient, createXrayClient } from './xray/index.ets';
```

## 功能使用

### 基础代理

```typescript
import { createXrayClient } from './xray';

// 创建客户端
const client = createXrayClient();

// 解析服务器配置 (VMess/VLESS/Trojan/SS)
const shareURL = "vmess://...";
const serverConfig = await client.parseShareURL(shareURL);

// 配置 Xray
const config = {
  inbound: {
    protocol: 'socks',
    port: 10808,
    listen: '127.0.0.1',
    settings: { auth: 'noauth', udp: true }
  },
  outbound: {
    protocol: 'vmess',
    settings: {
      vnext: [{
        address: serverConfig.address,
        port: serverConfig.port,
        users: [{
          id: serverConfig.id,
          alterId: serverConfig.alterId,
          security: serverConfig.security
        }]
      }]
    }
  }
};

// 启动
await client.loadConfig(config);
await client.start();

// 检查状态
console.log('Running:', client.isRunning());

// 获取统计
const stats = await client.getStats();
console.log('Stats:', stats);

// 停止
await client.stop();
client.destroy();
```

### 使用配置构建器

```typescript
import { ConfigBuilder } from './xray';

const builder = new ConfigBuilder();

// 设置日志
builder.setLogLevel('warning');

// 添加 SOCKS5 入站
builder.addSocksInbound(10808, '127.0.0.1', false, true);

// 添加 VMess 出站
builder.addVMessOutbound(
  'server.com',
  443,
  'uuid-string',
  0,
  'auto'
);

// 添加路由规则 (使用 geoip/geosite)
builder.addRoutingRule('field', 'direct', ['geosite:cn'], ['geoip:cn']);

// 启用统计
builder.enableStats();

// 构建配置
const config = builder.build();

// 使用配置
await client.loadConfig(config);
await client.start();
```

### 资产管理

```typescript
import { AssetManager } from './xray';

const assetMgr = new AssetManager('/data/storage/el2/base/assets');

// 检查资产状态
const geoipInfo = await assetMgr.getAssetInfo('geoip');
console.log('GeoIP exists:', geoipInfo.exists);

// 检查更新
const needsUpdate = await assetMgr.checkUpdate('geoip');
if (needsUpdate) {
  // 下载资产 (使用默认 URL)
  await assetMgr.download('geoip', '', (progress) => {
    console.log(`下载进度: ${progress.percentage}%`);
  });
}

// 验证资产
const valid = await assetMgr.verify('geoip');
console.log('GeoIP valid:', valid);
```

## VPN 实现

HarmonyOS 的 VPN 功能需要结合 VPN Extension Ability 和 Tun2Socks。

### 架构

```
HarmonyOS App
    ↓
VPN Extension Ability (创建 TUN 设备)
    ↓
Tun2Socks (处理 TUN 流量)
    ↓
Xray SOCKS5 (代理)
    ↓
远程服务器
```

### 实现步骤

#### 1. 配置 VPN Extension Ability

在 `module.json5` 中:

```json5
{
  "extensionAbilities": [
    {
      "name": "VPNExtension",
      "srcEntry": "./ets/vpnability/VPNExtensionAbility.ets",
      "type": "vpnExtension",
      "exported": true
    }
  ]
}
```

#### 2. 创建 VPN Extension

```typescript
// VPNExtensionAbility.ets
import vpnExtension from '@ohos.net.vpnExtension';

export default class VPNExtensionAbility extends vpnExtension.VpnExtensionAbility {
  onCreate(want: Want) {
    console.log('VPN Extension Created');
  }

  onConnect(want: Want) {
    console.log('VPN Extension Connected');
  }

  onDisconnect(want: Want) {
    console.log('VPN Extension Disconnected');
  }

  onDestroy() {
    console.log('VPN Extension Destroyed');
  }
}
```

#### 3. 设置 VPN 连接

```typescript
import vpnExtension from '@ohos.net.vpnExtension';
import { Tun2Socks, XrayClient } from './xray';

class VPNService {
  private vpnConnection: vpnExtension.VpnConnection;
  private xrayClient: XrayClient;
  private tun2socks: Tun2Socks;

  async setupVPN() {
    // 1. 创建 VPN 连接
    this.vpnConnection = vpnExtension.createVpnConnection(getContext());

    // 2. 配置 VPN
    const config: vpnExtension.VpnConfig = {
      addresses: [{ address: { address: '10.0.0.2' }, prefixLength: 24 }],
      routes: [{ interface: 'tun0', destination: { address: '0.0.0.0' }, prefixLength: 0 }],
      dnsServers: ['8.8.8.8', '8.8.4.4'],
      mtu: 1500,
    };

    // 3. 建立 VPN
    await this.vpnConnection.setUp(config);

    // 4. 获取 TUN 文件描述符
    const tunFd = this.vpnConnection.getFileDescriptor();

    // 5. 启动 Xray SOCKS5 代理
    this.xrayClient = createXrayClient();
    await this.xrayClient.loadConfig({
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
    await this.xrayClient.start();

    // 6. 启动 Tun2Socks
    this.tun2socks = new Tun2Socks({
      tunFd: tunFd,
      socksAddr: '127.0.0.1:10808',
      mtu: 1500,
      dnsAddr: '8.8.8.8:53'
    });
    await this.tun2socks.start();

    console.log('VPN 已启动');
  }

  async stopVPN() {
    // 停止 Tun2Socks
    if (this.tun2socks) {
      await this.tun2socks.stop();
    }

    // 停止 Xray
    if (this.xrayClient) {
      await this.xrayClient.stop();
      this.xrayClient.destroy();
    }

    // 断开 VPN 连接
    if (this.vpnConnection) {
      await this.vpnConnection.destroy();
    }

    console.log('VPN 已停止');
  }

  async getVPNStats() {
    const xrayStats = await this.xrayClient.getStats();
    const tunStats = await this.tun2socks.getStats();

    return {
      xray: xrayStats,
      tunnel: tunStats
    };
  }
}
```

#### 4. UI 控制

```typescript
// Index.ets
import { VPNService } from './VPNService';

@Entry
@Component
struct VPNControlPage {
  @State isConnected: boolean = false;
  private vpnService: VPNService = new VPNService();

  build() {
    Column() {
      Text(this.isConnected ? 'VPN 已连接' : 'VPN 未连接')
        .fontSize(24)
        .margin({ bottom: 20 })

      Button(this.isConnected ? '断开 VPN' : '连接 VPN')
        .onClick(async () => {
          if (this.isConnected) {
            await this.vpnService.stopVPN();
            this.isConnected = false;
          } else {
            await this.vpnService.setupVPN();
            this.isConnected = true;
          }
        })

      if (this.isConnected) {
        Button('查看统计')
          .onClick(async () => {
            const stats = await this.vpnService.getVPNStats();
            console.log('VPN Stats:', stats);
          })
      }
    }
    .width('100%')
    .height('100%')
    .justifyContent(FlexAlign.Center)
  }
}
```

### VPN 权限

在 `module.json5` 中添加必要权限:

```json5
{
  "requestPermissions": [
    {
      "name": "ohos.permission.INTERNET"
    },
    {
      "name": "ohos.permission.VPN_SETUP"
    }
  ]
}
```

## 常见问题

### Q: 编译失败,提示找不到 Xray-core 依赖?

**A**: 运行 `make install` 安装 Go 依赖:
```bash
cd go
go mod download
go mod tidy
```

### Q: 如何支持更多架构?

**A**: 修改 Makefile 的 `ARCHS` 变量:
```makefile
ARCHS := arm64 amd64 arm arm64-darwin
```

### Q: HarmonyOS 中如何加载 .so 库?

**A**: 使用 `import` 语句:
```typescript
import libxray from 'libxray.so';
```

### Q: Tun2Socks 性能如何优化?

**A**:
1. 调整 MTU 大小 (通常 1500 或 1420)
2. 使用合适的 DNS 服务器
3. 启用 UDP 支持
4. 考虑使用 FakeDNS

### Q: 如何调试 Xray 连接问题?

**A**:
1. 设置日志级别为 `debug`
2. 检查服务器配置是否正确
3. 测试 SOCKS5 代理是否可用
4. 查看 HarmonyOS 日志: `hdc shell hilog`

### Q: 支持哪些 Xray 协议?

**A**: 支持所有 Xray-core v1.25+ 支持的协议:
- VMess
- VLESS
- Trojan
- Shadowsocks
- Socks
- HTTP
- 以及所有传输层协议 (TCP, WS, gRPC, H2, QUIC等)

### Q: geoip/geosite 文件多大?

**A**:
- geoip.dat: ~4-6 MB
- geosite.dat: ~2-4 MB

总共约 6-10 MB,建议在 WiFi 下下载。

### Q: 如何实现开机自启动?

**A**: 使用 HarmonyOS 的后台任务能力:
1. 申请长时任务权限
2. 在后台服务中启动 VPN
3. 监听系统启动广播

### Q: 内存占用如何?

**A**:
- Xray 核心: ~20-50 MB
- Tun2Socks: ~10-20 MB
- 总计: ~30-70 MB (取决于配置和流量)

### Q: 支持 IPv6 吗?

**A**: 是的,Xray-core 完全支持 IPv6。在配置中正确设置即可。

## 性能调优

### 1. 日志级别

生产环境使用 `warning` 或 `error`:
```typescript
builder.setLogLevel('warning');
```

### 2. 启用多路复用

在 VMess/VLESS 配置中启用 mux:
```json
{
  "mux": {
    "enabled": true,
    "concurrency": 8
  }
}
```

### 3. DNS 配置

使用快速的 DNS 服务器:
```typescript
builder.setDNS(['223.5.5.5', '119.29.29.29'], {});
```

### 4. 路由优化

合理配置路由规则,避免不必要的代理:
```typescript
// 中国 IP/域名直连
builder.addRoutingRule('field', 'direct', ['geosite:cn'], ['geoip:cn', 'geoip:private']);

// 广告拦截
builder.addRoutingRule('field', 'block', ['geosite:category-ads-all'], []);
```

## 示例项目

完整的示例项目请参考:
- `examples/basic_usage.ets` - 基础使用示例
- `examples/VPNControl_Demo/` - 完整的 VPN 应用示例

## 技术支持

- GitHub Issues: https://github.com/shuffleman/XrayHarmony/issues
- 完整 API 文档: [XRAY_WRAPPER_API.md](./XRAY_WRAPPER_API.md)
- Xray 官方文档: https://xtls.github.io/

## 更新日志

### v2.0.0 (2025-12-06)
- ✨ 新增协议解析和生成工具 (VMess/VLESS/Trojan/SS)
- ✨ 新增 Tun2Socks 封装
- ✨ 新增资产管理器 (geoip/geosite)
- ✨ 新增配置构建器
- 🔧 增强流量统计功能
- 📚 完善文档

### v1.0.0 (2024-12-01)
- 🎉 初始版本
- ✅ 基础 Xray-core 封装
- ✅ Go -> C++ -> ArkTS 多层架构
