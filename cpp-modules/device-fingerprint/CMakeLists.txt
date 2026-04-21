# CMakeLists.txt for Device Fingerprint Module
# 高考志愿填报系统 - 设备指纹采集模块
# @author 高考志愿填报系统开发团队
# @version 1.0.0
# @date 2025-01-18

cmake_minimum_required(VERSION 3.16)

# 项目配置
project(DeviceFingerprint
    VERSION 1.0.0
    DESCRIPTION "High-performance cross-platform device fingerprint collection module"
    LANGUAGES CXX C
)

# C++标准设置
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# 编译选项配置
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

# 编译器特定选项
if(MSVC)
    # Windows/MSVC选项
    add_compile_options(
        /W4                     # 警告级别4
        /WX                     # 警告视为错误
        /permissive-           # 严格标准模式
        /Zc:__cplusplus        # 正确的__cplusplus宏
        /utf-8                 # UTF-8源文件编码
    )

    # Release优化选项
    if(CMAKE_BUILD_TYPE STREQUAL "Release")
        add_compile_options(
            /O2                 # 最大优化速度
            /Ob2                # 内联函数展开
            /Ot                 # 优化速度
            /Oy                 # 省略帧指针
            /GL                 # 全程序优化
        )
        add_link_options(/LTCG) # 链接时代码生成
    endif()

    # 安全编译选项
    add_compile_options(
        /GS                     # 缓冲区安全检查
        /sdl                    # SDL检查
    )

    # 预处理器定义
    add_compile_definitions(
        _CRT_SECURE_NO_WARNINGS # 禁用CRT安全警告
        WIN32_LEAN_AND_MEAN     # 减少Windows头文件
        NOMINMAX                # 禁用min/max宏
        UNICODE                 # Unicode支持
        _UNICODE
    )

else()
    # GCC/Clang选项
    add_compile_options(
        -Wall                   # 所有警告
        -Wextra                 # 额外警告
        -Werror                 # 警告视为错误
        -Wno-unused-parameter   # 忽略未使用参数警告
        -fPIC                   # 位置无关代码
    )

    # Release优化选项
    if(CMAKE_BUILD_TYPE STREQUAL "Release")
        add_compile_options(
            -O3                 # 最大优化
            -DNDEBUG           # 禁用断言
            -fomit-frame-pointer # 省略帧指针
            -ffast-math        # 快速数学
        )
    endif()

    # Debug选项
    if(CMAKE_BUILD_TYPE STREQUAL "Debug")
        add_compile_options(
            -g                  # 调试信息
            -O0                 # 无优化
            -fsanitize=address  # 地址检查器
            -fsanitize=undefined # 未定义行为检查器
        )
        add_link_options(-fsanitize=address -fsanitize=undefined)
    endif()

    # 安全编译选项
    add_compile_options(
        -fstack-protector-strong # 栈保护
        -D_FORTIFY_SOURCE=2     # 缓冲区溢出检查
    )

    # 位置无关可执行文件
    add_link_options(-pie)
endif()

# 平台特定配置
if(WIN32)
    set(PLATFORM_LIBS
        ws2_32
        iphlpapi
        wbemuuid
        ole32
        oleaut32
        psapi
    )
elseif(UNIX AND NOT APPLE)
    set(PLATFORM_LIBS
        pthread
        dl
    )

    # Linux特定库
    find_library(RT_LIB rt)
    if(RT_LIB)
        list(APPEND PLATFORM_LIBS ${RT_LIB})
    endif()

elseif(APPLE)
    set(PLATFORM_LIBS
        "-framework CoreFoundation"
        "-framework IOKit"
        "-framework Security"
    )
endif()

# 第三方依赖查找
find_package(PkgConfig REQUIRED)

# OpenSSL查找和配置
find_package(OpenSSL REQUIRED)
if(OpenSSL_FOUND)
    message(STATUS "OpenSSL version: ${OPENSSL_VERSION}")
    message(STATUS "OpenSSL include: ${OPENSSL_INCLUDE_DIR}")
    message(STATUS "OpenSSL libraries: ${OPENSSL_LIBRARIES}")
else()
    message(FATAL_ERROR "OpenSSL not found! Please install OpenSSL development packages.")
endif()

# Google Test查找(用于单元测试)
option(BUILD_TESTS "Build unit tests" ON)
if(BUILD_TESTS)
    find_package(GTest)
    if(NOT GTest_FOUND)
        # 如果找不到系统安装的GTest，尝试使用FetchContent下载
        include(FetchContent)
        FetchContent_Declare(
            googletest
            GIT_REPOSITORY https://github.com/google/googletest.git
            GIT_TAG release-1.12.1
        )
        FetchContent_MakeAvailable(googletest)
    endif()
endif()

# 包含目录
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${OPENSSL_INCLUDE_DIR}
)

# 源文件定义
set(HEADER_FILES
    include/device_fingerprint.h
    include/crypto_utils.h
    include/platform_detector.h
)

set(SOURCE_FILES
    src/device_fingerprint.cpp
    src/crypto_utils.cpp
    src/platform_detector.cpp
    src/c_interface.cpp
)

# 静态库目标
add_library(device_fingerprint_static STATIC
    ${HEADER_FILES}
    ${SOURCE_FILES}
)

# 动态库目标
add_library(device_fingerprint_shared SHARED
    ${HEADER_FILES}
    ${SOURCE_FILES}
)

# 库属性设置
set_target_properties(device_fingerprint_static PROPERTIES
    OUTPUT_NAME device_fingerprint
    CLEAN_DIRECT_OUTPUT 1
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
)

set_target_properties(device_fingerprint_shared PROPERTIES
    OUTPUT_NAME device_fingerprint
    CLEAN_DIRECT_OUTPUT 1
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
)

# 链接库
target_link_libraries(device_fingerprint_static
    PRIVATE
        OpenSSL::SSL
        OpenSSL::Crypto
        ${PLATFORM_LIBS}
)

target_link_libraries(device_fingerprint_shared
    PRIVATE
        OpenSSL::SSL
        OpenSSL::Crypto
        ${PLATFORM_LIBS}
)

# C接口导出(仅动态库)
if(WIN32)
    set_target_properties(device_fingerprint_shared PROPERTIES
        WINDOWS_EXPORT_ALL_SYMBOLS ON
    )
endif()

# 编译定义
target_compile_definitions(device_fingerprint_static
    PRIVATE
        DEVICE_FINGERPRINT_STATIC_LIB=1
)

target_compile_definitions(device_fingerprint_shared
    PRIVATE
        DEVICE_FINGERPRINT_SHARED_LIB=1
    PUBLIC
        DEVICE_FINGERPRINT_DLL=1
)

# 包含目录设置
target_include_directories(device_fingerprint_static
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

target_include_directories(device_fingerprint_shared
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# 示例程序
option(BUILD_EXAMPLES "Build example programs" ON)
if(BUILD_EXAMPLES)
    add_executable(fingerprint_example
        examples/fingerprint_example.cpp
    )

    target_link_libraries(fingerprint_example
        PRIVATE
            device_fingerprint_static
    )

    # C接口示例
    add_executable(c_interface_example
        examples/c_interface_example.c
    )

    target_link_libraries(c_interface_example
        PRIVATE
            device_fingerprint_shared
    )

    set_target_properties(c_interface_example PROPERTIES
        C_STANDARD 99
        C_STANDARD_REQUIRED ON
    )
endif()

# 单元测试
if(BUILD_TESTS)
    enable_testing()

    add_executable(device_fingerprint_tests
        tests/test_device_fingerprint.cpp
    )

    if(GTest_FOUND)
        target_link_libraries(device_fingerprint_tests
            PRIVATE
                device_fingerprint_static
                GTest::gtest
                GTest::gtest_main
                GTest::gmock
        )
    else()
        target_link_libraries(device_fingerprint_tests
            PRIVATE
                device_fingerprint_static
                gtest
                gtest_main
                gmock
        )
    endif()

    # 添加测试
    add_test(
        NAME DeviceFingerprintUnitTests
        COMMAND device_fingerprint_tests
    )

    # 设置测试环境
    set_tests_properties(DeviceFingerprintUnitTests
        PROPERTIES
            TIMEOUT 300
            ENVIRONMENT "GTEST_OUTPUT=xml:${CMAKE_BINARY_DIR}/test_results.xml"
    )
endif()

# 性能基准测试
option(BUILD_BENCHMARKS "Build performance benchmarks" OFF)
if(BUILD_BENCHMARKS)
    find_package(benchmark QUIET)
    if(benchmark_FOUND)
        add_executable(device_fingerprint_benchmark
            benchmarks/fingerprint_benchmark.cpp
        )

        target_link_libraries(device_fingerprint_benchmark
            PRIVATE
                device_fingerprint_static
                benchmark::benchmark
        )

        add_test(
            NAME DeviceFingerprintBenchmarks
            COMMAND device_fingerprint_benchmark
        )
    else()
        message(WARNING "Google Benchmark not found, skipping benchmark build")
    endif()
endif()

# Go绑定生成
option(BUILD_GO_BINDINGS "Generate Go language bindings" ON)
if(BUILD_GO_BINDINGS)
    # 创建Go模块文件
    configure_file(
        ${CMAKE_CURRENT_SOURCE_DIR}/bindings/go/device_fingerprint.go.in
        ${CMAKE_CURRENT_BINARY_DIR}/bindings/go/device_fingerprint.go
        @ONLY
    )

    # 生成CGO头文件
    add_custom_target(go_bindings
        COMMAND ${CMAKE_COMMAND} -E copy_if_different
                ${CMAKE_CURRENT_SOURCE_DIR}/include/device_fingerprint.h
                ${CMAKE_CURRENT_BINARY_DIR}/bindings/go/device_fingerprint.h
        DEPENDS device_fingerprint_shared
    )
endif()

# 安装配置
include(GNUInstallDirs)

# 安装头文件
install(
    FILES ${HEADER_FILES}
    DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/device_fingerprint
    COMPONENT Development
)

# 安装库文件
install(
    TARGETS device_fingerprint_static device_fingerprint_shared
    EXPORT DeviceFingerprintTargets
    ARCHIVE DESTINATION ${CMAKE_INSTALL_LIBDIR}
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
    COMPONENT Runtime
)

# 安装示例程序
if(BUILD_EXAMPLES)
    install(
        TARGETS fingerprint_example c_interface_example
        DESTINATION ${CMAKE_INSTALL_BINDIR}
        COMPONENT Examples
    )
endif()

# 生成和安装CMake配置文件
install(
    EXPORT DeviceFingerprintTargets
    FILE DeviceFingerprintTargets.cmake
    NAMESPACE DeviceFingerprint::
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/DeviceFingerprint
    COMPONENT Development
)

# 生成配置文件
include(CMakePackageConfigHelpers)

configure_package_config_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/cmake/DeviceFingerprintConfig.cmake.in
    ${CMAKE_CURRENT_BINARY_DIR}/DeviceFingerprintConfig.cmake
    INSTALL_DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/DeviceFingerprint
)

write_basic_package_version_file(
    ${CMAKE_CURRENT_BINARY_DIR}/DeviceFingerprintConfigVersion.cmake
    VERSION ${PROJECT_VERSION}
    COMPATIBILITY SameMajorVersion
)

install(
    FILES
        ${CMAKE_CURRENT_BINARY_DIR}/DeviceFingerprintConfig.cmake
        ${CMAKE_CURRENT_BINARY_DIR}/DeviceFingerprintConfigVersion.cmake
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/DeviceFingerprint
    COMPONENT Development
)

# pkg-config文件
configure_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/cmake/device_fingerprint.pc.in
    ${CMAKE_CURRENT_BINARY_DIR}/device_fingerprint.pc
    @ONLY
)

install(
    FILES ${CMAKE_CURRENT_BINARY_DIR}/device_fingerprint.pc
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/pkgconfig
    COMPONENT Development
)

# CPack配置
set(CPACK_PACKAGE_NAME "DeviceFingerprint")
set(CPACK_PACKAGE_VENDOR "高考志愿填报系统开发团队")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "High-performance cross-platform device fingerprint collection module")
set(CPACK_PACKAGE_VERSION_MAJOR ${PROJECT_VERSION_MAJOR})
set(CPACK_PACKAGE_VERSION_MINOR ${PROJECT_VERSION_MINOR})
set(CPACK_PACKAGE_VERSION_PATCH ${PROJECT_VERSION_PATCH})
set(CPACK_PACKAGE_INSTALL_DIRECTORY "DeviceFingerprint")
set(CPACK_RESOURCE_FILE_LICENSE "${CMAKE_CURRENT_SOURCE_DIR}/LICENSE")
set(CPACK_RESOURCE_FILE_README "${CMAKE_CURRENT_SOURCE_DIR}/README.md")

# 平台特定打包配置
if(WIN32)
    set(CPACK_GENERATOR "NSIS;ZIP")
    set(CPACK_NSIS_DISPLAY_NAME "Device Fingerprint Module")
    set(CPACK_NSIS_PACKAGE_NAME "DeviceFingerprint")
    set(CPACK_NSIS_CONTACT "support@gaokao-system.com")
elseif(APPLE)
    set(CPACK_GENERATOR "DragNDrop;TGZ")
else()
    set(CPACK_GENERATOR "DEB;RPM;TGZ")

    # Debian包配置
    set(CPACK_DEBIAN_PACKAGE_MAINTAINER "高考志愿填报系统开发团队 <dev@gaokao-system.com>")
    set(CPACK_DEBIAN_PACKAGE_DEPENDS "libssl-dev, libgtest-dev")
    set(CPACK_DEBIAN_PACKAGE_SECTION "devel")

    # RPM包配置
    set(CPACK_RPM_PACKAGE_GROUP "Development/Libraries")
    set(CPACK_RPM_PACKAGE_REQUIRES "openssl-devel, gtest-devel")
    set(CPACK_RPM_PACKAGE_LICENSE "Proprietary")
endif()

include(CPack)

# 自定义目标

# 代码格式化
find_program(CLANG_FORMAT clang-format)
if(CLANG_FORMAT)
    add_custom_target(format
        COMMAND ${CLANG_FORMAT} -i -style=file
                ${SOURCE_FILES} ${HEADER_FILES}
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
        COMMENT "Formatting source code"
    )
endif()

# 静态分析
find_program(CLANG_TIDY clang-tidy)
if(CLANG_TIDY)
    add_custom_target(static-analysis
        COMMAND ${CLANG_TIDY}
                ${SOURCE_FILES}
                -p ${CMAKE_BINARY_DIR}
                --config-file=${CMAKE_CURRENT_SOURCE_DIR}/.clang-tidy
        WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}
        COMMENT "Running static analysis"
    )
endif()

# 文档生成
find_package(Doxygen)
if(DOXYGEN_FOUND)
    configure_file(
        ${CMAKE_CURRENT_SOURCE_DIR}/docs/Doxyfile.in
        ${CMAKE_CURRENT_BINARY_DIR}/Doxyfile
        @ONLY
    )

    add_custom_target(docs
        COMMAND ${DOXYGEN_EXECUTABLE} ${CMAKE_CURRENT_BINARY_DIR}/Doxyfile
        WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
        COMMENT "Generating API documentation"
    )
endif()

# 清理目标
add_custom_target(distclean
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_BINARY_DIR}/CMakeFiles
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_BINARY_DIR}/CMakeCache.txt
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_BINARY_DIR}/cmake_install.cmake
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_BINARY_DIR}/Makefile
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    COMMENT "Cleaning build directory"
)

# 信息输出
message(STATUS "")
message(STATUS "Device Fingerprint Module Configuration Summary:")
message(STATUS "  Version:          ${PROJECT_VERSION}")
message(STATUS "  Build Type:       ${CMAKE_BUILD_TYPE}")
message(STATUS "  Platform:         ${CMAKE_SYSTEM_NAME}")
message(STATUS "  Architecture:     ${CMAKE_SYSTEM_PROCESSOR}")
message(STATUS "  Compiler:         ${CMAKE_CXX_COMPILER_ID} ${CMAKE_CXX_COMPILER_VERSION}")
message(STATUS "  C++ Standard:     ${CMAKE_CXX_STANDARD}")
message(STATUS "  OpenSSL Version:  ${OPENSSL_VERSION}")
message(STATUS "  Install Prefix:   ${CMAKE_INSTALL_PREFIX}")
message(STATUS "")
message(STATUS "Build Options:")
message(STATUS "  Build Tests:      ${BUILD_TESTS}")
message(STATUS "  Build Examples:   ${BUILD_EXAMPLES}")
message(STATUS "  Build Benchmarks: ${BUILD_BENCHMARKS}")
message(STATUS "  Build Go Bindings:${BUILD_GO_BINDINGS}")
message(STATUS "")

# 编译后信息
add_custom_target(build-info ALL
    COMMAND ${CMAKE_COMMAND} -E echo "Device Fingerprint Module build completed successfully!"
    COMMAND ${CMAKE_COMMAND} -E echo "Static Library: ${CMAKE_BINARY_DIR}/libdevice_fingerprint.a"
    COMMAND ${CMAKE_COMMAND} -E echo "Shared Library: ${CMAKE_BINARY_DIR}/libdevice_fingerprint.so"
    VERBATIM
)
