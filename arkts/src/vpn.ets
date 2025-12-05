/**
 * XrayHarmony VPN 模块
 * 提供 VPN 功能的 TypeScript/ArkTS 接口
 */

import libxray from 'libxray.so';

/**
 * VPN 配置接口
 */
export interface VPNConfig {
  /** TUN 设备文件描述符 */
  tunFd: number;

  /** TUN MTU（最大传输单元） */
  tunMTU?: number;

  /** SOCKS5 代理地址（连接到 Xray） */
  socksAddr: string;

  /** DNS 服务器列表 */
  dnsServers?: string[];

  /** 是否启用 FakeDNS */
  fakeDNS?: boolean;

  /** 是否启用 UDP */
  udp?: boolean;

  /** 是否启用 TCP 并发 */
  tcpConcurrent?: boolean;
}

/**
 * VPN 统计信息接口
 */
export interface VPNStats {
  /** 是否正在运行 */
  running: boolean;

  /** SOCKS 地址 */
  socksAddr: string;

  /** MTU 值 */
  mtu: number;
}

/**
 * Xray VPN 客户端
 * 结合 Xray 和 TUN 网卡实现 VPN 功能
 */
export class XrayVPNClient {
  private xrayInstanceId: number = -1;
  private vpnManagerId: number = -1;
  private native: any;

  /**
   * 构造函数
   * @param xrayInstanceId Xray 实例 ID（来自 XrayClient）
   */
  constructor(xrayInstanceId: number) {
    this.native = libxray;
    this.xrayInstanceId = xrayInstanceId;

    // 创建 VPN 管理器
    this.vpnManagerId = this.native.VPNNewManager(xrayInstanceId);
    if (this.vpnManagerId < 0) {
      throw new Error(`Failed to create VPN manager: ${this.getLastError()}`);
    }
  }

  /**
   * 启动 VPN
   * @param config VPN 配置
   */
  async start(config: VPNConfig): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      // 设置默认值
      const fullConfig: Required<VPNConfig> = {
        tunFd: config.tunFd,
        tunMTU: config.tunMTU || 1400,
        socksAddr: config.socksAddr,
        dnsServers: config.dnsServers || ['8.8.8.8', '8.8.4.4'],
        fakeDNS: config.fakeDNS !== undefined ? config.fakeDNS : false,
        udp: config.udp !== undefined ? config.udp : true,
        tcpConcurrent: config.tcpConcurrent !== undefined ? config.tcpConcurrent : false
      };

      // 转换为 JSON 字符串
      const configJSON = JSON.stringify(fullConfig);

      // 调用 native 方法
      const result = this.native.VPNStart(this.vpnManagerId, configJSON);
      if (result !== 0) {
        reject(new Error(`Failed to start VPN: ${this.getLastError()}`));
      } else {
        resolve();
      }
    });
  }

  /**
   * 停止 VPN
   */
  async stop(): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const result = this.native.VPNStop(this.vpnManagerId);
      if (result !== 0) {
        reject(new Error(`Failed to stop VPN: ${this.getLastError()}`));
      } else {
        resolve();
      }
    });
  }

  /**
   * 检查 VPN 是否正在运行
   */
  isRunning(): boolean {
    const result = this.native.VPNIsRunning(this.vpnManagerId);
    return result === 1;
  }

  /**
   * 获取 VPN 统计信息
   */
  async getStats(): Promise<VPNStats> {
    return new Promise<VPNStats>((resolve, reject) => {
      const statsJSON = this.native.VPNGetStats(this.vpnManagerId);
      if (!statsJSON) {
        reject(new Error(`Failed to get stats: ${this.getLastError()}`));
        return;
      }

      try {
        const stats = JSON.parse(statsJSON);
        // 释放 native 分配的字符串
        this.native.XrayFreeString(statsJSON);
        resolve(stats);
      } catch (error) {
        reject(new Error(`Failed to parse stats: ${error}`));
      }
    });
  }

  /**
   * 获取最后的错误信息
   */
  getLastError(): string {
    const error = this.native.XrayGetLastError();
    if (error) {
      const errorStr = error.toString();
      this.native.XrayFreeString(error);
      return errorStr;
    }
    return '';
  }

  /**
   * 销毁 VPN 客户端
   */
  destroy(): void {
    if (this.vpnManagerId >= 0) {
      if (this.isRunning()) {
        this.native.VPNStop(this.vpnManagerId);
      }
      this.native.VPNDeleteManager(this.vpnManagerId);
      this.vpnManagerId = -1;
    }
  }
}

/**
 * 创建 XrayVPN 客户端
 * @param xrayInstanceId Xray 实例 ID
 */
export function createXrayVPNClient(xrayInstanceId: number): XrayVPNClient {
  return new XrayVPNClient(xrayInstanceId);
}
