# CMakeLists.txt for Volunteer Matcher C++ Module

cmake_minimum_required(VERSION 3.12)
project(VolunteerMatcher)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 设置ONNX Runtime路径（根据实际安装路径调整）
set(ONNXRUNTIME_ROOT_DIR "/usr/local/onnxruntime")
set(ONNXRUNTIME_INCLUDE_DIRS ${ONNXRUNTIME_ROOT_DIR}/include)
set(ONNXRUNTIME_LIBRARIES ${ONNXRUNTIME_ROOT_DIR}/lib/libonnxruntime.so)

# 设置JsonCpp路径
find_package(PkgConfig REQUIRED)
pkg_check_modules(JSONCPP REQUIRED jsoncpp)

# 包含目录
include_directories(
    ${PROJECT_SOURCE_DIR}/include
    ${ONNXRUNTIME_INCLUDE_DIRS}
    ${JSONCPP_INCLUDE_DIRS}
)

# 源文件
file(GLOB SOURCES "src/*.cpp")

# 显式启用ONNX Runtime
add_definitions(-DENABLE_ONNX_RUNTIME)

# 创建共享库
add_library(volunteer_matcher SHARED ${SOURCES})

# 链接库
target_link_libraries(volunteer_matcher 
    ${ONNXRUNTIME_LIBRARIES}
    ${JSONCPP_LIBRARIES}
)

# 安装规则
install(TARGETS volunteer_matcher
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)

install(DIRECTORY include/ 
    DESTINATION include
    FILES_MATCHING PATTERN "*.h"
)