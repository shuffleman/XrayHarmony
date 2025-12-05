/**
 * Type definitions for XrayHarmony
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

export interface XrayStats {
  running: boolean;
  status: string;
  uptime?: number;
  traffic?: {
    uplink: number;
    downlink: number;
  };
}

export class XrayClient {
  constructor();

  loadConfig(config: XrayConfig): Promise<void>;
  loadConfigFromFile(filePath: string): Promise<void>;
  testConfig(config: XrayConfig): Promise<boolean>;

  start(): Promise<void>;
  stop(): Promise<void>;
  isRunning(): boolean;

  getStats(): Promise<XrayStats>;
  getLastError(): string;

  static getVersion(): string;
  destroy(): void;
}

export function createXrayClient(): XrayClient;
export const VERSION: string;
