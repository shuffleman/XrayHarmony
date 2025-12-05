.PHONY: all clean build-go build-native help test

# 默认目标
all: build-go build-native

# 显示帮助信息
help:
	@echo "XrayHarmony - Build System"
	@echo "=========================="
	@echo ""
	@echo "Available targets:"
	@echo "  all          - Build everything (default)"
	@echo "  build-go     - Build Go shared libraries"
	@echo "  build-native - Build native C++ wrapper"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  install      - Install dependencies"
	@echo ""
	@echo "Build for specific architecture:"
	@echo "  make build-go ARCH=arm64"
	@echo "  make build-go ARCH=amd64"
	@echo "  make build-go ARCH=arm"

# 架构参数（默认为 all）
ARCH ?= all

# 构建 Go 共享库
build-go:
	@echo "Building Go shared libraries..."
	chmod +x build/build.sh
	./build/build.sh $(ARCH)

# 构建 Native C++ 包装器
build-native:
	@echo "Building native C++ wrapper..."
	mkdir -p build/cmake
	cd build/cmake && cmake .. && make
	@echo "Native build completed"

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	rm -rf libs/*
	rm -rf build/cmake
	cd go && go clean
	@echo "Clean completed"

# 安装依赖
install:
	@echo "Installing dependencies..."
	cd go && go mod download
	@echo "Dependencies installed"

# 运行测试
test:
	@echo "Running tests..."
	cd go && go test -v ./...
	@echo "Tests completed"

# 初始化项目
init: install
	@echo "Initializing project..."
	mkdir -p libs
	mkdir -p build/cmake
	@echo "Project initialized"
