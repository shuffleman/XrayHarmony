# Xray-core 升级计划

## 当前状态

- **Xray-core**: v1.8.16
- **Go 版本**: 1.24.7
- **gvisor**: v0.0.0-20231202080848-1f7806d17489 (旧版本，与 v1.8.16 兼容)

## 升级目标

升级到 **Xray-core v1.250911.0**（2025.09.11 版本）

## 前置要求

Xray-core v1.250911.0 需要 **Go 1.25+**

## 升级步骤

### 1. 升级 Go 工具链到 1.25.5

在 GitHub Actions 或有网络连接的环境中：

```bash
cd go
go get golang.org/toolchain@go1.25.5
```

或修改 `go.mod`:

```go
toolchain go1.25.5
```

### 2. 升级 Xray-core

```bash
go get github.com/xtls/xray-core@v1.250911.0
go mod tidy
```

### 3. 验证编译

```bash
go build -buildmode=c-shared -o libxray.so ./wrapper/
```

### 4. 检查 gvisor 版本

新版本的 Xray-core 可能已经更新了 gvisor 依赖，检查是否与新版 gvisor 兼容：

```bash
go list -m all | grep gvisor
```

如果 gvisor 版本仍然冲突，可能需要：
- 等待 Xray-core 更新 WireGuard 模块
- 或者在构建时排除 WireGuard 代理

## 预期好处

1. **最新功能**: 获得 Xray-core 的最新特性和优化
2. **安全更新**: 包含最新的安全补丁
3. **依赖更新**: 可能已解决 gvisor 版本冲突
4. **性能提升**: 新版本通常包含性能优化

## 风险评估

- **中等风险**: API 可能有变化，需要测试兼容性
- **构建环境**: 需要 CI/CD 环境支持 Go 1.25+
- **依赖冲突**: 可能引入新的依赖冲突

## 回滚计划

如果升级失败，可以回滚到 v1.8.16：

```bash
go get github.com/xtls/xray-core@v1.8.16
go get gvisor.dev/gvisor@v0.0.0-20231202080848-1f7806d17489
go mod tidy
```

## 参考

- [Xray-core Releases](https://github.com/XTLS/Xray-core/releases)
- [Go Downloads](https://go.dev/dl/)
