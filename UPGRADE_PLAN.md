# Xray-core 升级记录

## ✅ 升级已完成

升级已于 2025年9月成功完成，并持续保持更新至最新稳定版本。

## 当前状态

- **Xray-core**: v1.251202.0 (2025年12月2日版本)
- **Go 版本**: 1.25
- **Go 工具链**: go1.25.5
- **gvisor**: v0.0.0-20250428193742-2d800c3129d5 (最新版本)

## 升级历史

### 第二次升级: v1.251202.0 (2025年12月)

**升级时间**: 2025年12月

**变更内容**:
- ✅ Xray-core 升级到 v1.251202.0
- ✅ 保持 Go 1.25 工具链
- ✅ gvisor 依赖已完全兼容
- ✅ 所有功能正常工作

### 第一次重大升级: v1.250911.0 (2025年9月)

**升级时间**: 2025年9月

**变更内容**:
- ✅ Go 工具链从 1.24.7 升级到 1.25.5
- ✅ Xray-core 从 v1.8.16 升级到 v1.250911.0
- ✅ gvisor 依赖冲突已解决
- ✅ 移除了嵌入式 tun2socks 的尝试，明确分离式架构

**关键成果**:
1. **依赖冲突解决**: 成功解决了 gvisor 版本冲突问题
2. **代码重构**: 移除了嵌入式 tun2socks 的尝试，明确架构方向
3. **核心功能稳定**: Xray-core 代理功能完全正常工作
4. **性能提升**: 新版本带来显著的性能改进
5. **安全加固**: 包含最新的安全补丁

**架构说明**:
- XrayHarmony 专注于封装 Xray-core 核心功能
- Xray-core 本身不支持 TUN 入站
- VPN 功能需要配合外部 tun2socks 组件实现
- 采用分离式架构确保模块清晰和稳定性

## 技术细节

### Go 工具链升级

```bash
# 在 go.mod 中设置
go 1.25
toolchain go1.25.5
```

### Xray-core 依赖

```go
require github.com/xtls/xray-core v1.251202.0
```

### 关键依赖更新

- `gvisor.dev/gvisor v0.0.0-20250428193742-2d800c3129d5`
- `golang.zx2c4.com/wireguard v0.0.0-20231211153847-12269c276173`
- `github.com/quic-go/quic-go v0.57.1`
- `golang.org/x/crypto v0.44.0`

## 兼容性验证

所有核心功能已验证工作正常：
- ✅ 基础代理功能（SOCKS5、HTTP 等）
- ✅ 多协议支持（VMess、VLESS、Trojan、Shadowsocks 等）
- ✅ 流量路由和分流
- ✅ DNS 处理
- ⚠️ VPN 功能需配合外部 tun2socks 使用（参见 VPN_ARCHITECTURE.md）

## 未来维护

### 保持更新策略

1. **定期检查**: 每月检查 Xray-core 新版本
2. **安全优先**: 有安全更新时立即升级
3. **功能更新**: 根据需求评估功能更新
4. **依赖管理**: 保持所有依赖在最新稳定版本

### 升级流程

```bash
# 1. 检查新版本
cd go
go list -m -u github.com/xtls/xray-core

# 2. 升级到新版本
go get github.com/xtls/xray-core@latest
go mod tidy

# 3. 验证编译
go build -buildmode=c-shared -o libxray.so ./wrapper/

# 4. 测试功能
make test
```

## 回滚信息

如果需要回滚到特定版本：

```bash
# 回滚到 v1.250911.0
cd go
go get github.com/xtls/xray-core@v1.250911.0
go mod tidy

# 或回滚到 v1.8.16（原始版本）
go get github.com/xtls/xray-core@v1.8.16
go get gvisor.dev/gvisor@v0.0.0-20231202080848-1f7806d17489
go mod tidy
```

## 相关文档

- [构建指南](docs/BUILD.md) - 包含 Go 1.25+ 的构建说明
- [API 文档](docs/API.md) - API 接口说明
- [VPN 文档](docs/VPN.md) - VPN 功能使用指南

## 参考资源

- [Xray-core Releases](https://github.com/XTLS/Xray-core/releases)
- [Go Downloads](https://go.dev/dl/)
- [gvisor Releases](https://github.com/google/gvisor/releases)
