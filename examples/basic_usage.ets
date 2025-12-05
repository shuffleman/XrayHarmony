/**
 * XrayHarmony Basic Usage Example
 *
 * This example demonstrates how to use XrayHarmony in a HarmonyOS application.
 */

import { XrayClient, XrayConfig } from '@shuffleman/xray-harmony';

/**
 * 基础使用示例
 */
export async function basicUsageExample() {
  // 创建 Xray 客户端实例
  const client = new XrayClient();

  try {
    // 配置 Xray
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
      },
      log: {
        loglevel: 'info'
      }
    };

    // 加载配置
    console.info('Loading configuration...');
    await client.loadConfig(config);

    // 启动 Xray
    console.info('Starting Xray...');
    await client.start();

    // 检查运行状态
    if (client.isRunning()) {
      console.info('Xray is running!');

      // 获取统计信息
      const stats = await client.getStats();
      console.info('Stats:', JSON.stringify(stats));
    }

    // 运行一段时间后停止
    setTimeout(async () => {
      console.info('Stopping Xray...');
      await client.stop();
      console.info('Xray stopped');

      // 释放资源
      client.destroy();
    }, 30000); // 30 秒后停止

  } catch (error) {
    console.error('Error:', error);
    const lastError = client.getLastError();
    if (lastError) {
      console.error('Last error:', lastError);
    }
    client.destroy();
  }
}

/**
 * 从文件加载配置的示例
 */
export async function loadConfigFromFileExample() {
  const client = new XrayClient();

  try {
    // 从文件加载配置
    const configPath = '/data/storage/el2/base/xray_config.json';
    console.info('Loading configuration from file...');
    await client.loadConfigFromFile(configPath);

    // 启动 Xray
    await client.start();
    console.info('Xray started successfully');

    // ... 使用 Xray

    // 停止并清理
    await client.stop();
    client.destroy();

  } catch (error) {
    console.error('Error:', error);
    client.destroy();
  }
}

/**
 * 配置测试示例
 */
export async function testConfigExample() {
  const client = new XrayClient();

  try {
    const config: XrayConfig = {
      inbound: {
        protocol: 'socks',
        port: 1080,
        listen: '127.0.0.1'
      },
      outbound: {
        protocol: 'freedom'
      }
    };

    // 测试配置是否有效
    console.info('Testing configuration...');
    const isValid = await client.testConfig(config);

    if (isValid) {
      console.info('Configuration is valid');
      // 继续加载和启动
      await client.loadConfig(config);
      await client.start();
    } else {
      console.error('Configuration is invalid');
    }

  } catch (error) {
    console.error('Error:', error);
  } finally {
    client.destroy();
  }
}

/**
 * 完整的应用集成示例
 */
export class XrayService {
  private client: XrayClient | null = null;
  private isInitialized: boolean = false;

  async initialize(config: XrayConfig): Promise<boolean> {
    try {
      this.client = new XrayClient();
      await this.client.loadConfig(config);
      this.isInitialized = true;
      return true;
    } catch (error) {
      console.error('Failed to initialize XrayService:', error);
      return false;
    }
  }

  async start(): Promise<boolean> {
    if (!this.isInitialized || !this.client) {
      console.error('XrayService not initialized');
      return false;
    }

    try {
      await this.client.start();
      return true;
    } catch (error) {
      console.error('Failed to start Xray:', error);
      return false;
    }
  }

  async stop(): Promise<boolean> {
    if (!this.client) {
      return true;
    }

    try {
      await this.client.stop();
      return true;
    } catch (error) {
      console.error('Failed to stop Xray:', error);
      return false;
    }
  }

  isRunning(): boolean {
    return this.client?.isRunning() ?? false;
  }

  async getStats() {
    if (!this.client) {
      return null;
    }

    try {
      return await this.client.getStats();
    } catch (error) {
      console.error('Failed to get stats:', error);
      return null;
    }
  }

  destroy(): void {
    if (this.client) {
      this.client.destroy();
      this.client = null;
      this.isInitialized = false;
    }
  }
}
