cmake_minimum_required(VERSION 3.16)
project(XrayHarmony)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 包含目录
include_directories(
    ${CMAKE_SOURCE_DIR}/../native/include
)
include_directories(${CMAKE_BINARY_DIR}/libs)

# 库目录
link_directories(
    ${CMAKE_SOURCE_DIR}/../libs
)

# 源文件
set(SOURCES
    ${CMAKE_SOURCE_DIR}/../native/src/xray_bridge.cpp
)

# 创建共享库
add_library(xray_harmony SHARED ${SOURCES})

# 链接 Go 共享库
# 根据目标平台选择正确的库
if(CMAKE_SYSTEM_PROCESSOR MATCHES "aarch64" OR CMAKE_SYSTEM_PROCESSOR MATCHES "arm64")
    set(XRAY_LIB "xray_linux_arm64")
elseif(CMAKE_SYSTEM_PROCESSOR MATCHES "x86_64" OR CMAKE_SYSTEM_PROCESSOR MATCHES "amd64")
    set(XRAY_LIB "xray_linux_amd64")
elseif(CMAKE_SYSTEM_PROCESSOR MATCHES "arm")
    set(XRAY_LIB "xray_linux_arm")
else()
    message(FATAL_ERROR "Unsupported architecture: ${CMAKE_SYSTEM_PROCESSOR}")
endif()

target_link_libraries(xray_harmony ${XRAY_LIB})

# 安装规则
install(TARGETS xray_harmony
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
)

install(DIRECTORY ${CMAKE_SOURCE_DIR}/../native/include/
    DESTINATION include/xray_harmony
    FILES_MATCHING PATTERN "*.h"
)
