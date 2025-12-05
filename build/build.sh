#!/bin/bash

# XrayHarmony Build Script
# This script builds the Go shared library for HarmonyOS

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}XrayHarmony Build Script${NC}"
echo "================================"

# 检查环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_DIR="$PROJECT_ROOT/go"
OUTPUT_DIR="$PROJECT_ROOT/libs"

echo "Project root: $PROJECT_ROOT"
echo "Go directory: $GO_DIR"
echo "Output directory: $OUTPUT_DIR"

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 支持的目标平台
TARGETS=(
    "linux/arm64"
    "linux/amd64"
    "linux/arm"
)

# 构建函数
build_for_target() {
    local os=$1
    local arch=$2
    local output_name="libxray_${os}_${arch}.so"

    echo -e "\n${YELLOW}Building for ${os}/${arch}...${NC}"

    cd "$GO_DIR"

    # 设置环境变量
    export GOOS=$os
    export GOARCH=$arch
    export CGO_ENABLED=1

    # 根据架构设置 CC
    case "$arch" in
        arm64)
            export CC=aarch64-linux-gnu-gcc
            ;;
        arm)
            export CC=arm-linux-gnueabihf-gcc
            export GOARM=7
            ;;
        amd64)
            export CC=gcc
            ;;
    esac

    # 构建共享库
    go build -buildmode=c-shared -o "$OUTPUT_DIR/$output_name" ./wrapper

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built: $output_name${NC}"
        # 生成文件信息
        ls -lh "$OUTPUT_DIR/$output_name"
    else
        echo -e "${RED}✗ Failed to build: $output_name${NC}"
        return 1
    fi
}

# 解析命令行参数
TARGET_ARCH="${1:-all}"

if [ "$TARGET_ARCH" = "all" ]; then
    echo -e "\n${YELLOW}Building for all targets...${NC}"
    for target in "${TARGETS[@]}"; do
        IFS='/' read -r os arch <<< "$target"
        build_for_target "$os" "$arch" || true
    done
else
    echo -e "\n${YELLOW}Building for specific target: $TARGET_ARCH${NC}"
    case "$TARGET_ARCH" in
        arm64)
            build_for_target "linux" "arm64"
            ;;
        amd64|x86_64)
            build_for_target "linux" "amd64"
            ;;
        arm)
            build_for_target "linux" "arm"
            ;;
        *)
            echo -e "${RED}Unknown target: $TARGET_ARCH${NC}"
            echo "Usage: $0 [all|arm64|amd64|arm]"
            exit 1
            ;;
    esac
fi

echo -e "\n${GREEN}Build completed!${NC}"
echo "Output directory: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR"
