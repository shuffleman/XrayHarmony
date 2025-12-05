/**
 * XrayHarmony VPN 服务示例
 * 展示如何使用 VpnExtensionAbility 和 XrayVPN 实现完整的 VPN 功能
 */

import VpnExtensionAbility from '@ohos.app.ability.VpnExtensionAbility';
import vpnExt from '@ohos.net.vpnExtension';
import { XrayClient, createXrayClient } from '../arkts/src/index';
import { XrayVPNClient, createXrayVPNClient, VPNConfig } from '../arkts/src/vpn';

/**
 * VPN 扩展能力服务
 * 需要在 module.json5 中配置：
 * {
 *   "extensionAbilities": [{
 *     "name": "XrayVpnExtension",
 *     "srcEntry": "./ets/vpnextension/XrayVpnExtension.ets",
 *     "type": "vpn.extension"
 *   }]
 * }
 */
export default class XrayVpnExtension extends VpnExtensionAbility {
  private xrayClient: XrayClient | null = null;
  private vpnClient: XrayVPNClient | null = null;
  private vpnConnection: vpnExt.VpnConnection | null = null;
  private tunFd: number = -1;

  /**
   * 扩展能力创建时的回调
   */
  onCreate(): void {
    console.info('[XrayVPN] VPN Extension created');

    try {
      // 1. 创建 VPN 连接对象
      this.vpnConnection = vpnExt.createVpnConnection(this.context);
      console.info('[XrayVPN] VPN connection created');
    } catch (error) {
      console.error(`[XrayVPN] Failed to create VPN connection: ${error}`);
    }
  }

  /**
   * 启动 VPN 服务
   */
  async startVPN(xrayConfig: any, vpnConfig?: Partial<VPNConfig>): Promise<void> {
    try {
      // 1. 创建 Xray 客户端
      console.info('[XrayVPN] Creating Xray client...');
      this.xrayClient = createXrayClient();

      // 2. 加载 Xray 配置
      console.info('[XrayVPN] Loading Xray config...');
      await this.xrayClient.loadConfig(xrayConfig);

      // 3. 启动 Xray（建立 SOCKS5 入站）
      console.info('[XrayVPN] Starting Xray...');
      await this.xrayClient.start();

      // 4. 创建 TUN 虚拟网卡
      console.info('[XrayVPN] Creating TUN device...');
      const tunConfig: vpnExt.VpnConfig = {
        // VPN 虚拟网卡 IP 地址和前缀长度
        addresses: [
          {
            address: {
              address: '10.0.0.2',
              family: 1  // IPv4
            },
            prefixLength: 24
          }
        ],

        // 路由配置
        routes: [
          {
            interface: 'vpn-tun',
            destination: {
              address: '0.0.0.0',
              family: 1
            },
            prefixLength: 0,
            gateway: {
              address: '10.0.0.1',
              family: 1
            }
          }
        ],

        // MTU
        mtu: vpnConfig?.tunMTU || 1400,

        // DNS 服务器
        dnsAddresses: vpnConfig?.dnsServers?.map(dns => ({
          address: dns,
          family: 1
        })) || [
          { address: '8.8.8.8', family: 1 },
          { address: '8.8.4.4', family: 1 }
        ],

        // 应用白名单/黑名单（可选）
        // trustedApplications: [],
        // blockedApplications: []
      };

      // 创建 TUN 设备并获取文件描述符
      this.tunFd = await this.vpnConnection!.create(tunConfig);
      console.info(`[XrayVPN] TUN device created, fd: ${this.tunFd}`);

      // 5. 创建 VPN 客户端并启动
      console.info('[XrayVPN] Creating VPN client...');
      const xrayInstanceId = (this.xrayClient as any).instanceId; // 获取 Xray 实例 ID
      this.vpnClient = createXrayVPNClient(xrayInstanceId);

      // 构建完整的 VPN 配置
      const fullVpnConfig: VPNConfig = {
        tunFd: this.tunFd,
        tunMTU: vpnConfig?.tunMTU || 1400,
        socksAddr: vpnConfig?.socksAddr || '127.0.0.1:10808', // Xray SOCKS5 入站地址
        dnsServers: vpnConfig?.dnsServers || ['8.8.8.8', '8.8.4.4'],
        fakeDNS: vpnConfig?.fakeDNS || false,
        udp: vpnConfig?.udp !== undefined ? vpnConfig.udp : true,
        tcpConcurrent: vpnConfig?.tcpConcurrent || false
      };

      console.info('[XrayVPN] Starting VPN with config:', JSON.stringify(fullVpnConfig));
      await this.vpnClient.start(fullVpnConfig);

      console.info('[XrayVPN] VPN started successfully');
    } catch (error) {
      console.error(`[XrayVPN] Failed to start VPN: ${error}`);
      await this.stopVPN();
      throw error;
    }
  }

  /**
   * 停止 VPN 服务
   */
  async stopVPN(): Promise<void> {
    console.info('[XrayVPN] Stopping VPN...');

    try {
      // 1. 停止 VPN 客户端
      if (this.vpnClient) {
        await this.vpnClient.stop();
        this.vpnClient.destroy();
        this.vpnClient = null;
      }

      // 2. 销毁 TUN 设备
      if (this.vpnConnection && this.tunFd >= 0) {
        await this.vpnConnection.destroy();
        this.tunFd = -1;
      }

      // 3. 停止 Xray
      if (this.xrayClient) {
        await this.xrayClient.stop();
        this.xrayClient.destroy();
        this.xrayClient = null;
      }

      console.info('[XrayVPN] VPN stopped successfully');
    } catch (error) {
      console.error(`[XrayVPN] Failed to stop VPN: ${error}`);
      throw error;
    }
  }

  /**
   * 获取 VPN 状态
   */
  async getStatus(): Promise<any> {
    if (!this.vpnClient || !this.xrayClient) {
      return {
        vpnRunning: false,
        xrayRunning: false
      };
    }

    try {
      const vpnStats = await this.vpnClient.getStats();
      const xrayStats = await this.xrayClient.getStats();

      return {
        vpnRunning: this.vpnClient.isRunning(),
        xrayRunning: this.xrayClient.isRunning(),
        vpnStats,
        xrayStats
      };
    } catch (error) {
      console.error(`[XrayVPN] Failed to get status: ${error}`);
      return {
        vpnRunning: false,
        xrayRunning: false,
        error: error.toString()
      };
    }
  }

  /**
   * 扩展能力销毁时的回调
   */
  onDestroy(): void {
    console.info('[XrayVPN] VPN Extension destroying...');

    // 停止 VPN（异步，但不等待）
    this.stopVPN().catch(error => {
      console.error(`[XrayVPN] Error during destruction: ${error}`);
    });
  }
}
