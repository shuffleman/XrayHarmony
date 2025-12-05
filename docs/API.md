# XrayHarmony API 文档

## 概述

XrayHarmony 为鸿蒙系统提供了完整的 Xray-core 封装，支持在 HarmonyOS 应用中使用 Xray 的所有核心功能。

## 架构层次

```
┌─────────────────────────────────┐
│   HarmonyOS Application (ArkTS)  │
├─────────────────────────────────┤
│   XrayHarmony ArkTS Interface   │
├─────────────────────────────────┤
│   Native C++ Bridge Layer       │
├─────────────────────────────────┤
│   Go Wrapper (CGO)              │
├─────────────────────────────────┤
│   Xray-core                     │
└─────────────────────────────────┘
```

## ArkTS API

### XrayClient 类

主要的客户端类，用于管理 Xray 实例。

#### 构造函数

```typescript
constructor()
```

创建一个新的 XrayClient 实例。

**异常：**
- 如果无法创建实例，将抛出 Error

**示例：**
```typescript
const client = new XrayClient();
```

#### loadConfig

```typescript
async loadConfig(config: XrayConfig): Promise<void>
```

从配置对象加载 Xray 配置。

**参数：**
- `config: XrayConfig` - Xray 配置对象

**返回：**
- `Promise<void>` - 配置加载成功时 resolve

**异常：**
- 配置无效或加载失败时抛出 Error

**示例：**
```typescript
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

await client.loadConfig(config);
```

#### loadConfigFromFile

```typescript
async loadConfigFromFile(filePath: string): Promise<void>
```

从文件加载 Xray 配置。

**参数：**
- `filePath: string` - 配置文件的完整路径

**返回：**
- `Promise<void>` - 配置加载成功时 resolve

**异常：**
- 文件不存在或配置无效时抛出 Error

**示例：**
```typescript
await client.loadConfigFromFile('/data/storage/el2/base/config.json');
```

#### testConfig

```typescript
async testConfig(config: XrayConfig): Promise<boolean>
```

测试配置是否有效，不会实际应用配置。

**参数：**
- `config: XrayConfig` - 要测试的配置对象

**返回：**
- `Promise<boolean>` - 配置有效返回 true

**异常：**
- 配置测试失败时抛出 Error

**示例：**
```typescript
const isValid = await client.testConfig(config);
if (isValid) {
  await client.loadConfig(config);
}
```

#### start

```typescript
async start(): Promise<void>
```

启动 Xray 实例。必须先调用 `loadConfig` 或 `loadConfigFromFile`。

**返回：**
- `Promise<void>` - 启动成功时 resolve

**异常：**
- 未加载配置或启动失败时抛出 Error

**示例：**
```typescript
await client.start();
console.log('Xray started');
```

#### stop

```typescript
async stop(): Promise<void>
```

停止 Xray 实例。

**返回：**
- `Promise<void>` - 停止成功时 resolve

**异常：**
- 停止失败时抛出 Error

**示例：**
```typescript
await client.stop();
console.log('Xray stopped');
```

#### isRunning

```typescript
isRunning(): boolean
```

检查 Xray 是否正在运行。

**返回：**
- `boolean` - 正在运行返回 true，否则返回 false

**示例：**
```typescript
if (client.isRunning()) {
  console.log('Xray is running');
}
```

#### getStats

```typescript
async getStats(): Promise<XrayStats>
```

获取 Xray 运行统计信息。

**返回：**
- `Promise<XrayStats>` - 统计信息对象

**异常：**
- 实例未运行或获取失败时抛出 Error

**示例：**
```typescript
const stats = await client.getStats();
console.log('Running:', stats.running);
console.log('Status:', stats.status);
```

#### getLastError

```typescript
getLastError(): string
```

获取最后一次操作的错误信息。

**返回：**
- `string` - 错误信息，无错误时返回空字符串

**示例：**
```typescript
try {
  await client.start();
} catch (error) {
  console.error('Error:', client.getLastError());
}
```

#### destroy

```typescript
destroy(): void
```

释放资源，清理 Xray 实例。应在不再使用时调用。

**示例：**
```typescript
client.destroy();
```

#### getVersion (静态方法)

```typescript
static getVersion(): string
```

获取 XrayHarmony 的版本信息。

**返回：**
- `string` - 版本字符串

**示例：**
```typescript
const version = XrayClient.getVersion();
console.log('Version:', version);
```

## 类型定义

### XrayConfig

```typescript
interface XrayConfig {
  inbound?: InboundConfig;
  outbound?: OutboundConfig;
  log?: LogConfig;
}
```

### InboundConfig

```typescript
interface InboundConfig {
  protocol: string;      // 协议类型：socks, http, vmess, vless 等
  port: number;          // 监听端口
  listen?: string;       // 监听地址，默认 127.0.0.1
  settings?: object;     // 协议特定设置
}
```

### OutboundConfig

```typescript
interface OutboundConfig {
  protocol: string;      // 协议类型：freedom, blackhole, vmess, vless 等
  settings?: object;     // 协议特定设置
}
```

### LogConfig

```typescript
interface LogConfig {
  loglevel?: 'debug' | 'info' | 'warning' | 'error' | 'none';
}
```

### XrayStats

```typescript
interface XrayStats {
  running: boolean;      // 是否正在运行
  status: string;        // 状态描述
  uptime?: number;       // 运行时间（秒）
  traffic?: {            // 流量统计
    uplink: number;      // 上行流量（字节）
    downlink: number;    // 下行流量（字节）
  };
}
```

## 工具函数

### createXrayClient

```typescript
function createXrayClient(): XrayClient
```

便捷的工厂函数，创建并返回新的 XrayClient 实例。

**示例：**
```typescript
import { createXrayClient } from '@shuffleman/xray-harmony';

const client = createXrayClient();
```

## 最佳实践

### 1. 资源管理

始终在使用完毕后调用 `destroy()` 释放资源：

```typescript
const client = new XrayClient();
try {
  await client.loadConfig(config);
  await client.start();
  // ... 使用 Xray
} finally {
  client.destroy();
}
```

### 2. 错误处理

正确处理异步操作的错误：

```typescript
try {
  await client.start();
} catch (error) {
  console.error('Failed to start:', error);
  const lastError = client.getLastError();
  console.error('Details:', lastError);
}
```

### 3. 配置验证

在应用配置前先测试：

```typescript
if (await client.testConfig(config)) {
  await client.loadConfig(config);
} else {
  console.error('Invalid configuration');
}
```

### 4. 状态检查

在执行操作前检查状态：

```typescript
if (!client.isRunning()) {
  await client.start();
}
```

## 示例应用

完整的示例代码请参见 `examples/` 目录：

- `basic_usage.ets` - 基础使用示例
- `config.json` - 配置文件示例

## 注意事项

1. **权限要求**：应用需要网络权限才能使用 Xray 功能
2. **线程安全**：XrayClient 实例不是线程安全的，应在单一线程中使用
3. **资源限制**：注意系统资源限制，特别是文件描述符数量
4. **配置安全**：敏感配置信息应妥善保护，不要硬编码在代码中

## 故障排除

### 常见问题

**Q: 启动失败，提示"config not loaded"**

A: 确保在调用 `start()` 前先调用 `loadConfig()` 或 `loadConfigFromFile()`

**Q: 如何查看详细日志？**

A: 在配置中设置 `log.loglevel` 为 `debug`

**Q: 性能问题或内存泄漏**

A: 确保在不使用时调用 `destroy()` 释放资源

## 支持

如有问题或建议，请访问：
https://github.com/shuffleman/XrayHarmony/issues
