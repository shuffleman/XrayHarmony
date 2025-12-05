/**
 * XrayHarmony - Xray-core wrapper for HarmonyOS
 *
 * This module provides ArkTS/TypeScript interface for Xray-core functionality
 * on HarmonyOS platform.
 */

import libxray from 'libxray.so';

/**
 * Xray 配置接口
 */
export interface XrayConfig {
  inbound?: InboundConfig;
  outbound?: OutboundConfig;
  log?: LogConfig;
}

export interface InboundConfig {
  protocol: string;
  port: number;
  listen?: string;
  settings?: object;
}

export interface OutboundConfig {
  protocol: string;
  settings?: object;
}

export interface LogConfig {
  loglevel?: 'debug' | 'info' | 'warning' | 'error' | 'none';
}

/**
 * Xray 统计信息接口
 */
export interface XrayStats {
  running: boolean;
  status: string;
  uptime?: number;
  traffic?: {
    uplink: number;
    downlink: number;
  };
}

/**
 * XrayClient 类 - 主要的 Xray 客户端接口
 */
export class XrayClient {
  private instanceId: number = -1;
  private native: any;

  constructor() {
    this.native = libxray;
    this.instanceId = this.native.XrayNewInstance();

    if (this.instanceId < 0) {
      throw new Error('Failed to create Xray instance');
    }
  }

  /**
   * 加载 JSON 配置
   * @param config - Xray 配置对象
   * @returns Promise<void>
   */
  async loadConfig(config: XrayConfig): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        const configJSON = JSON.stringify(config);
        const result = this.native.XrayLoadConfig(this.instanceId, configJSON);

        if (result !== 0) {
          const error = this.getLastError();
          reject(new Error(error || 'Failed to load config'));
        } else {
          resolve();
        }
      } catch (err) {
        reject(err);
      }
    });
  }

  /**
   * 从文件加载配置
   * @param filePath - 配置文件路径
   * @returns Promise<void>
   */
  async loadConfigFromFile(filePath: string): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        const result = this.native.XrayLoadConfigFromFile(this.instanceId, filePath);

        if (result !== 0) {
          const error = this.getLastError();
          reject(new Error(error || 'Failed to load config from file'));
        } else {
          resolve();
        }
      } catch (err) {
        reject(err);
      }
    });
  }

  /**
   * 测试配置是否有效
   * @param config - Xray 配置对象
   * @returns Promise<boolean>
   */
  async testConfig(config: XrayConfig): Promise<boolean> {
    return new Promise((resolve, reject) => {
      try {
        const configJSON = JSON.stringify(config);
        const result = this.native.XrayTestConfig(this.instanceId, configJSON);

        if (result !== 0) {
          const error = this.getLastError();
          reject(new Error(error || 'Config test failed'));
        } else {
          resolve(true);
        }
      } catch (err) {
        reject(err);
      }
    });
  }

  /**
   * 启动 Xray 实例
   * @returns Promise<void>
   */
  async start(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        const result = this.native.XrayStart(this.instanceId);

        if (result !== 0) {
          const error = this.getLastError();
          reject(new Error(error || 'Failed to start Xray'));
        } else {
          resolve();
        }
      } catch (err) {
        reject(err);
      }
    });
  }

  /**
   * 停止 Xray 实例
   * @returns Promise<void>
   */
  async stop(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        const result = this.native.XrayStop(this.instanceId);

        if (result !== 0) {
          const error = this.getLastError();
          reject(new Error(error || 'Failed to stop Xray'));
        } else {
          resolve();
        }
      } catch (err) {
        reject(err);
      }
    });
  }

  /**
   * 检查 Xray 是否正在运行
   * @returns boolean
   */
  isRunning(): boolean {
    try {
      const result = this.native.XrayIsRunning(this.instanceId);
      return result === 1;
    } catch {
      return false;
    }
  }

  /**
   * 获取统计信息
   * @returns Promise<XrayStats>
   */
  async getStats(): Promise<XrayStats> {
    return new Promise((resolve, reject) => {
      try {
        const statsJSON = this.native.XrayGetStats(this.instanceId);

        if (!statsJSON) {
          const error = this.getLastError();
          reject(new Error(error || 'Failed to get stats'));
        } else {
          const stats = JSON.parse(statsJSON) as XrayStats;
          resolve(stats);
        }
      } catch (err) {
        reject(err);
      }
    });
  }

  /**
   * 获取最后的错误信息
   * @returns string
   */
  getLastError(): string {
    try {
      return this.native.XrayGetLastError() || '';
    } catch {
      return '';
    }
  }

  /**
   * 获取版本信息
   * @returns string
   */
  static getVersion(): string {
    try {
      return libxray.XrayGetVersion() || 'Unknown';
    } catch {
      return 'Unknown';
    }
  }

  /**
   * 释放资源
   */
  destroy(): void {
    if (this.instanceId >= 0) {
      if (this.isRunning()) {
        this.native.XrayStop(this.instanceId);
      }
      this.native.XrayDeleteInstance(this.instanceId);
      this.instanceId = -1;
    }
  }
}

/**
 * 便捷的工厂函数
 */
export function createXrayClient(): XrayClient {
  return new XrayClient();
}

/**
 * 导出版本信息
 */
export const VERSION = XrayClient.getVersion();
