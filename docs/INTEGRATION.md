# HarmonyOS 集成指南

本文档详细说明如何将 XrayHarmony 集成到 HarmonyOS 应用中。

## 前置条件

- HarmonyOS SDK (API 9 或更高)
- DevEco Studio
- 已编译的 XrayHarmony 库文件

## 集成步骤

### 1. 准备库文件

首先，构建适合你目标设备的库文件：

```bash
# 对于大多数鸿蒙设备（ARM64）
make build-go ARCH=arm64
```

这将在 `libs/` 目录生成：
- `libxray_linux_arm64.so` - 共享库
- `libxray_linux_arm64.h` - C 头文件

### 2. 创建 HarmonyOS 项目

在 DevEco Studio 中创建新的 HarmonyOS 应用项目。

### 3. 添加库文件

#### 3.1 复制共享库

将编译好的 `.so` 文件复制到项目的 Native 库目录：

```
entry/
  ├── src/
  │   └── main/
  │       ├── cpp/
  │       │   └── libs/
  │       │       └── arm64-v8a/
  │       │           └── libxray_linux_arm64.so
  │       └── ets/
```

#### 3.2 复制 ArkTS 接口文件

将 `arkts/src/` 目录下的文件复制到项目中：

```
entry/
  └── src/
      └── main/
          └── ets/
              └── xray/
                  ├── index.ets
                  └── index.d.ts
```

### 4. 配置项目

#### 4.1 修改 build-profile.json5

在 `entry/build-profile.json5` 中添加 Native 库配置：

```json
{
  "apiType": "stageMode",
  "buildOption": {
    "arkOptions": {
      "runtimeOnly": {
        "sources": [
          "./src/main/ets/xray"
        ]
      }
    },
    "externalNativeOptions": {
      "path": "./src/main/cpp/CMakeLists.txt",
      "arguments": "-DOHOS_STL=c++_shared",
      "cppFlags": "",
    }
  }
}
```

#### 4.2 创建 CMakeLists.txt

在 `entry/src/main/cpp/` 目录创建 `CMakeLists.txt`：

```cmake
cmake_minimum_required(VERSION 3.4.1)
project(XrayHarmonyApp)

set(NATIVERENDER_ROOT_PATH ${CMAKE_CURRENT_SOURCE_DIR})

# 添加预编译库
add_library(xray SHARED IMPORTED)
set_target_properties(xray PROPERTIES IMPORTED_LOCATION
    ${NATIVERENDER_ROOT_PATH}/libs/${OHOS_ARCH}/libxray_linux_arm64.so)

# 包含目录
include_directories(${NATIVERENDER_ROOT_PATH})
```

#### 4.3 配置模块权限

在 `entry/src/main/module.json5` 中添加必要权限：

```json
{
  "module": {
    "requestPermissions": [
      {
        "name": "ohos.permission.INTERNET"
      },
      {
        "name": "ohos.permission.GET_NETWORK_INFO"
      }
    ]
  }
}
```

### 5. 使用 XrayHarmony

#### 5.1 基础用法

在你的 ArkTS 代码中导入并使用：

```typescript
import { XrayClient, XrayConfig } from '../xray/index';

@Entry
@Component
struct Index {
  private xrayClient: XrayClient | null = null;

  aboutToAppear() {
    this.initXray();
  }

  aboutToDisappear() {
    if (this.xrayClient) {
      this.xrayClient.destroy();
    }
  }

  async initXray() {
    try {
      this.xrayClient = new XrayClient();

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

      await this.xrayClient.loadConfig(config);
      console.info('Xray initialized');
    } catch (error) {
      console.error('Failed to initialize Xray:', error);
    }
  }

  build() {
    Column() {
      Button('Start Xray')
        .onClick(async () => {
          try {
            await this.xrayClient?.start();
            console.info('Xray started');
          } catch (error) {
            console.error('Failed to start:', error);
          }
        })

      Button('Stop Xray')
        .onClick(async () => {
          try {
            await this.xrayClient?.stop();
            console.info('Xray stopped');
          } catch (error) {
            console.error('Failed to stop:', error);
          }
        })
    }
  }
}
```

#### 5.2 创建服务类

建议创建一个单独的服务类管理 Xray：

```typescript
// XrayService.ets
import { XrayClient, XrayConfig, XrayStats } from '../xray/index';

export class XrayService {
  private static instance: XrayService;
  private client: XrayClient | null = null;
  private _isRunning: boolean = false;

  private constructor() {}

  static getInstance(): XrayService {
    if (!XrayService.instance) {
      XrayService.instance = new XrayService();
    }
    return XrayService.instance;
  }

  async initialize(config: XrayConfig): Promise<boolean> {
    try {
      if (this.client) {
        this.client.destroy();
      }

      this.client = new XrayClient();
      await this.client.loadConfig(config);
      return true;
    } catch (error) {
      console.error('Failed to initialize:', error);
      return false;
    }
  }

  async start(): Promise<boolean> {
    if (!this.client) {
      console.error('Client not initialized');
      return false;
    }

    try {
      await this.client.start();
      this._isRunning = true;
      return true;
    } catch (error) {
      console.error('Failed to start:', error);
      return false;
    }
  }

  async stop(): Promise<boolean> {
    if (!this.client) {
      return true;
    }

    try {
      await this.client.stop();
      this._isRunning = false;
      return true;
    } catch (error) {
      console.error('Failed to stop:', error);
      return false;
    }
  }

  isRunning(): boolean {
    return this._isRunning && (this.client?.isRunning() ?? false);
  }

  async getStats(): Promise<XrayStats | null> {
    if (!this.client || !this._isRunning) {
      return null;
    }

    try {
      return await this.client.getStats();
    } catch (error) {
      console.error('Failed to get stats:', error);
      return null;
    }
  }

  destroy() {
    if (this.client) {
      if (this._isRunning) {
        this.client.stop().then(() => {
          this.client?.destroy();
          this.client = null;
          this._isRunning = false;
        });
      } else {
        this.client.destroy();
        this.client = null;
      }
    }
  }
}
```

#### 5.3 在 UIAbility 中使用

```typescript
// EntryAbility.ts
import UIAbility from '@ohos.app.ability.UIAbility';
import { XrayService } from '../ets/services/XrayService';

export default class EntryAbility extends UIAbility {
  onCreate(want, launchParam) {
    console.info('Ability onCreate');

    // 初始化 Xray
    this.initializeXray();
  }

  async initializeXray() {
    const xrayService = XrayService.getInstance();

    const config = {
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

    const success = await xrayService.initialize(config);
    if (success) {
      console.info('Xray service initialized');
    }
  }

  onDestroy() {
    console.info('Ability onDestroy');

    // 清理资源
    const xrayService = XrayService.getInstance();
    xrayService.destroy();
  }
}
```

### 6. 配置管理

#### 6.1 使用 Preferences 存储配置

```typescript
import dataPreferences from '@ohos.data.preferences';

export class ConfigManager {
  private preferences: dataPreferences.Preferences | null = null;

  async init(context) {
    this.preferences = await dataPreferences.getPreferences(context, 'xray_config');
  }

  async saveConfig(config: XrayConfig): Promise<void> {
    if (!this.preferences) {
      throw new Error('Preferences not initialized');
    }

    await this.preferences.put('config', JSON.stringify(config));
    await this.preferences.flush();
  }

  async loadConfig(): Promise<XrayConfig | null> {
    if (!this.preferences) {
      throw new Error('Preferences not initialized');
    }

    const configStr = await this.preferences.get('config', '');
    if (configStr) {
      return JSON.parse(configStr as string);
    }
    return null;
  }
}
```

#### 6.2 从文件加载配置

```typescript
import fs from '@ohos.file.fs';

async loadConfigFromFile(filePath: string): Promise<XrayConfig> {
  try {
    const file = fs.openSync(filePath, fs.OpenMode.READ_ONLY);
    const buffer = new ArrayBuffer(4096);
    const readLen = fs.readSync(file.fd, buffer);
    fs.closeSync(file);

    const decoder = new util.TextDecoder('utf-8');
    const configStr = decoder.decode(new Uint8Array(buffer, 0, readLen));
    return JSON.parse(configStr);
  } catch (error) {
    console.error('Failed to load config from file:', error);
    throw error;
  }
}
```

### 7. 错误处理

实现统一的错误处理：

```typescript
export class XrayError extends Error {
  constructor(public code: string, message: string) {
    super(message);
    this.name = 'XrayError';
  }
}

export async function safeXrayOperation<T>(
  operation: () => Promise<T>,
  errorMessage: string
): Promise<T> {
  try {
    return await operation();
  } catch (error) {
    console.error(errorMessage, error);
    throw new XrayError('OPERATION_FAILED', errorMessage);
  }
}

// 使用示例
await safeXrayOperation(
  () => xrayClient.start(),
  'Failed to start Xray'
);
```

### 8. 日志管理

```typescript
export class XrayLogger {
  private static readonly TAG = 'XrayHarmony';

  static info(message: string, ...args: any[]) {
    console.info(`[${this.TAG}] ${message}`, ...args);
  }

  static error(message: string, error?: any) {
    console.error(`[${this.TAG}] ${message}`, error);
  }

  static debug(message: string, ...args: any[]) {
    console.debug(`[${this.TAG}] ${message}`, ...args);
  }
}
```

### 9. 性能优化

#### 9.1 延迟初始化

```typescript
private xrayClient: XrayClient | null = null;

getXrayClient(): XrayClient {
  if (!this.xrayClient) {
    this.xrayClient = new XrayClient();
  }
  return this.xrayClient;
}
```

#### 9.2 资源释放

确保在适当的生命周期中释放资源：

```typescript
aboutToDisappear() {
  // 清理资源
  this.xrayClient?.destroy();
  this.xrayClient = null;
}
```

### 10. 测试

#### 10.1 单元测试

```typescript
import { describe, it, expect } from '@ohos/hypium';
import { XrayClient } from '../xray/index';

export default function XrayTest() {
  describe('XrayClient', () => {
    it('should create instance', () => {
      const client = new XrayClient();
      expect(client).not.toBeNull();
      client.destroy();
    });

    it('should load config', async () => {
      const client = new XrayClient();
      const config = {
        inbound: { protocol: 'socks', port: 1080 },
        outbound: { protocol: 'freedom' }
      };

      await expect(client.loadConfig(config)).resolves.toBeUndefined();
      client.destroy();
    });
  });
}
```

## 常见问题

### 1. 库加载失败

**问题**: `dlopen failed: library "libxray_linux_arm64.so" not found`

**解决方案**:
- 确保 .so 文件在正确的目录
- 检查架构是否匹配（arm64-v8a, armeabi-v7a）
- 验证 CMakeLists.txt 配置

### 2. 权限问题

**问题**: 网络操作失败

**解决方案**:
在 `module.json5` 中添加必要权限。

### 3. 内存泄漏

**问题**: 应用内存持续增长

**解决方案**:
确保调用 `destroy()` 释放资源。

## 最佳实践

1. **单例模式**: 使用单例管理 Xray 实例
2. **错误处理**: 所有异步操作都要有 try-catch
3. **资源管理**: 在组件销毁时清理资源
4. **配置安全**: 不要在代码中硬编码敏感信息
5. **日志记录**: 记录关键操作和错误

## 下一步

- 查看 [API 文档](API.md) 了解详细接口
- 参考 [示例代码](../examples/) 学习更多用法
- 阅读 [构建指南](BUILD.md) 了解如何构建库

## 支持

如有集成问题，请访问：
https://github.com/shuffleman/XrayHarmony/issues
