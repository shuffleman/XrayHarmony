# 构建指南

本文档详细说明如何构建 XrayHarmony 项目。

## 系统要求

### 必需软件

- **Go**: **1.25 或更高版本** (推荐 1.25.5+)
  - ⚠️ **重要**: Xray-core v1.251202.0 需要 Go 1.25+
  - 低于此版本将无法编译
- **CMake**: 3.16 或更高版本
- **Make**: GNU Make 或兼容版本
- **GCC/Clang**: 支持 C++17 的编译器

### 交叉编译工具链

根据目标平台，你可能需要以下工具链：

#### ARM64 (aarch64)
```bash
# Ubuntu/Debian
sudo apt-get install gcc-aarch64-linux-gnu g++-aarch64-linux-gnu

# macOS (使用 Homebrew)
brew install aarch64-elf-gcc
```

#### ARM (armv7)
```bash
# Ubuntu/Debian
sudo apt-get install gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf
```

#### AMD64/x86_64
通常系统自带的 GCC 即可。

## 快速构建

### 1. 克隆仓库

```bash
git clone https://github.com/shuffleman/XrayHarmony.git
cd XrayHarmony
```

### 2. 安装 Go 依赖

```bash
make install
```

这将下载所有必需的 Go 模块，包括 xray-core。

### 3. 构建所有目标

```bash
make all
```

这将构建所有支持的架构（ARM64、AMD64、ARM）。

### 4. 构建特定架构

```bash
# 只构建 ARM64
make build-go ARCH=arm64

# 只构建 AMD64
make build-go ARCH=amd64

# 只构建 ARM
make build-go ARCH=arm
```

## 详细构建步骤

### 第一步：Go 共享库

Go 共享库是项目的核心，包含 Xray-core 的封装。

```bash
./build/build.sh [all|arm64|amd64|arm]
```

**参数说明：**
- `all` (默认): 构建所有架构
- `arm64`: 只构建 ARM64
- `amd64`: 只构建 AMD64
- `arm`: 只构建 ARM

**输出：**
编译后的共享库将位于 `libs/` 目录：
```
libs/
├── libxray_linux_arm64.so
├── libxray_linux_arm64.h
├── libxray_linux_amd64.so
├── libxray_linux_amd64.h
├── libxray_linux_arm.so
└── libxray_linux_arm.h
```

### 第二步：Native C++ 桥接层（可选）

如果需要构建 C++ 桥接层：

```bash
make build-native
```

这将使用 CMake 构建 Native 层。

**输出：**
编译后的库将位于 `build/cmake/` 目录。

## 环境变量

### Go 构建环境变量

构建脚本会自动设置以下环境变量：

```bash
# 目标操作系统
export GOOS=linux

# 目标架构
export GOARCH=arm64  # 或 amd64, arm

# 启用 CGO
export CGO_ENABLED=1

# C 编译器（根据目标架构）
export CC=aarch64-linux-gnu-gcc  # ARM64
# 或
export CC=gcc  # AMD64
# 或
export CC=arm-linux-gnueabihf-gcc  # ARM
```

你可以在运行构建脚本前手动设置这些变量以自定义构建过程。

## 常见问题

### 1. 找不到交叉编译器

**错误信息：**
```
aarch64-linux-gnu-gcc: command not found
```

**解决方案：**
安装相应的交叉编译工具链（参见"系统要求"部分）。

### 2. Go 版本过低

**错误信息：**
```
go: github.com/xtls/xray-core@v1.251202.0 requires go >= 1.25
```

**解决方案：**
升级 Go 到 1.25 或更高版本：
```bash
# 方法1: 下载并安装 Go 1.25
# 访问 https://go.dev/dl/ 下载最新版本

# 方法2: 使用 Go 工具链管理（推荐）
cd go
go get golang.org/toolchain@go1.25.5
```

### 3. Go 模块下载失败

**错误信息：**
```
go: github.com/xtls/xray-core@v1.251202.0: Get "https://proxy.golang.org/...": dial tcp: i/o timeout
```

**解决方案：**
设置 Go 代理：
```bash
export GOPROXY=https://goproxy.cn,direct
# 或
export GOPROXY=https://goproxy.io,direct
```

### 4. CGO 编译错误

**错误信息：**
```
# runtime/cgo
gcc: error: unrecognized command line option '-marm64'
```

**解决方案：**
确保使用正确的交叉编译器，检查 `CC` 环境变量是否设置正确。

### 5. 权限错误

**错误信息：**
```
./build/build.sh: Permission denied
```

**解决方案：**
给脚本添加执行权限：
```bash
chmod +x build/build.sh
```

## 高级构建选项

### 自定义输出目录

```bash
OUTPUT_DIR=/custom/path ./build/build.sh
```

### 启用调试符号

修改 `build/build.sh`，在 `go build` 命令中添加 `-gcflags="all=-N -l"`:

```bash
go build -gcflags="all=-N -l" -buildmode=c-shared -o "$OUTPUT_DIR/$output_name" ./wrapper
```

### 静态链接

要创建静态链接的库，修改构建模式：

```bash
go build -buildmode=c-archive -o libxray.a ./wrapper
```

### 减小二进制大小

使用 `-ldflags` 参数去除调试信息：

```bash
go build -ldflags="-s -w" -buildmode=c-shared -o libxray.so ./wrapper
```

## 验证构建

### 检查共享库

```bash
# 查看库信息
file libs/libxray_linux_arm64.so

# 查看导出符号
nm -D libs/libxray_linux_arm64.so | grep Xray

# 检查依赖
ldd libs/libxray_linux_arm64.so
```

### 运行测试

```bash
make test
```

## 清理构建产物

```bash
# 清理所有构建产物
make clean

# 只清理 Go 构建缓存
cd go && go clean -cache
```

## 构建优化建议

### 1. 并行构建

如果有多核处理器，可以并行构建多个架构：

```bash
# 在不同终端中
make build-go ARCH=arm64 &
make build-go ARCH=amd64 &
make build-go ARCH=arm &
wait
```

### 2. 使用构建缓存

Go 会自动缓存构建结果，重复构建会更快。

### 3. 交叉编译最佳实践

- 使用相同版本的交叉编译工具链
- 确保 Go 版本与目标平台兼容
- 测试构建后的库是否能在目标平台运行

## 针对 HarmonyOS 的特殊说明

### 鸿蒙系统架构

鸿蒙系统主要运行在以下架构：
- **ARM64**: 大多数鸿蒙设备（手机、平板）
- **ARM**: 部分物联网设备

建议优先构建 ARM64 版本：

```bash
make build-go ARCH=arm64
```

### 鸿蒙 NDK

如果使用鸿蒙 NDK，需要设置：

```bash
export OHOS_NDK_HOME=/path/to/ohos-sdk/native
export CC=$OHOS_NDK_HOME/llvm/bin/clang
```

然后按正常流程构建。

## 持续集成 (CI)

### GitHub Actions 示例

```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        arch: [arm64, amd64, arm]

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Install cross-compilers
        run: |
          sudo apt-get update
          sudo apt-get install -y gcc-aarch64-linux-gnu gcc-arm-linux-gnueabihf

      - name: Build
        run: make build-go ARCH=${{ matrix.arch }}

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: libs-${{ matrix.arch }}
          path: libs/
```

## 发布

构建发布版本：

```bash
# 清理并重新构建
make clean
make all

# 验证构建
ls -lh libs/

# 创建发布包
tar -czf XrayHarmony-libs.tar.gz libs/
```

## 支持

如有构建问题，请：
1. 查看本文档的"常见问题"部分
2. 检查 [GitHub Issues](https://github.com/shuffleman/XrayHarmony/issues)
3. 提交新的 Issue 并附上详细的错误信息
