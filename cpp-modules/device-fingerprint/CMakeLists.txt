# CMakeLists.txt for Device Fingerprint C++ Module

cmake_minimum_required(VERSION 3.12)
project(DeviceFingerprint)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 设置包含目录
include_directories(
    ${PROJECT_SOURCE_DIR}/include
)

# 源文件
file(GLOB SOURCES "src/*.cpp")

# 创建共享库
add_library(device_fingerprint SHARED ${SOURCES})

# 链接库
target_link_libraries(device_fingerprint)

# 安装规则
install(TARGETS device_fingerprint
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)

install(DIRECTORY include/ 
    DESTINATION include
    FILES_MATCHING PATTERN "*.h"
)