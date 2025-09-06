<<<<<<< HEAD
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
=======
# CMakeLists.txt for Volunteer Matcher Module
# 高考志愿填报系统 - 志愿匹配算法引擎
# @author 高考志愿填报系统开发团队
# @version 1.0.0
# @date 2025-01-18

cmake_minimum_required(VERSION 3.16)

# 项目配置
project(VolunteerMatcher 
    VERSION 1.0.0
    DESCRIPTION "High-performance volunteer matching algorithm engine for Gaokao"
    LANGUAGES CXX
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
        /bigobj                # 大对象文件支持
    )
    
    # Release优化选项
    if(CMAKE_BUILD_TYPE STREQUAL "Release")
        add_compile_options(
            /O2                 # 最大优化速度
            /Ob2                # 内联函数展开
            /Ot                 # 优化速度
            /Oy                 # 省略帧指针
            /GL                 # 全程序优化
            /DNDEBUG           # 禁用断言
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
        _SILENCE_CXX17_ITERATOR_BASE_CLASS_DEPRECATION_WARNING
    )
    
else()
    # GCC/Clang选项
    add_compile_options(
        -Wall                   # 所有警告
        -Wextra                 # 额外警告
        -Werror                 # 警告视为错误
        -Wno-unused-parameter   # 忽略未使用参数警告
        -fPIC                   # 位置无关代码
        -march=native           # 针对当前CPU优化
    )
    
    # Release优化选项
    if(CMAKE_BUILD_TYPE STREQUAL "Release")
        add_compile_options(
            -O3                 # 最大优化
            -DNDEBUG           # 禁用断言
            -fomit-frame-pointer # 省略帧指针
            -ffast-math        # 快速数学
            -flto              # 链接时优化
        )
        add_link_options(-flto)
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
        advapi32
    )
elseif(UNIX AND NOT APPLE)
    set(PLATFORM_LIBS 
        pthread 
        dl
        m
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
        "-framework Accelerate"
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

# JsonCpp查找和配置
find_package(PkgConfig QUIET)
if(PkgConfig_FOUND)
    pkg_check_modules(JsonCpp jsoncpp)
endif()

if(NOT JsonCpp_FOUND)
    find_package(jsoncpp QUIET)
    if(jsoncpp_FOUND)
        set(JsonCpp_LIBRARIES jsoncpp_lib)
        set(JsonCpp_INCLUDE_DIRS ${jsoncpp_INCLUDE_DIRS})
    else()
        # 尝试手动查找
        find_path(JsonCpp_INCLUDE_DIRS json/json.h 
            PATHS /usr/include /usr/local/include /opt/local/include)
        find_library(JsonCpp_LIBRARIES NAMES jsoncpp libjsoncpp
            PATHS /usr/lib /usr/local/lib /opt/local/lib)
        
        if(NOT JsonCpp_INCLUDE_DIRS OR NOT JsonCpp_LIBRARIES)
            message(FATAL_ERROR "JsonCpp not found! Please install libjsoncpp-dev or jsoncpp development packages.")
        endif()
    endif()
endif()

message(STATUS "JsonCpp include: ${JsonCpp_INCLUDE_DIRS}")
message(STATUS "JsonCpp libraries: ${JsonCpp_LIBRARIES}")

# XGBoost查找 (可选)
find_package(xgboost QUIET)
if(xgboost_FOUND)
    message(STATUS "XGBoost found: ${xgboost_VERSION}")
    add_compile_definitions(HAVE_XGBOOST=1)
else()
    message(WARNING "XGBoost not found. Machine learning features will be limited.")
    add_compile_definitions(HAVE_XGBOOST=0)
endif()

# OpenMP查找 (用于并行计算)
find_package(OpenMP QUIET)
if(OpenMP_CXX_FOUND)
    message(STATUS "OpenMP found: ${OpenMP_CXX_VERSION}")
else()
    message(WARNING "OpenMP not found. Parallel computation will be limited.")
endif()

# Google Test查找(用于单元测试)
option(BUILD_TESTS "Build unit tests" ON)
if(BUILD_TESTS)
    find_package(GTest QUIET)
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

# Google Benchmark查找(用于性能测试)
option(BUILD_BENCHMARKS "Build performance benchmarks" OFF)
if(BUILD_BENCHMARKS)
    find_package(benchmark QUIET)
    if(NOT benchmark_FOUND)
        message(WARNING "Google Benchmark not found, skipping benchmark build")
        set(BUILD_BENCHMARKS OFF)
    endif()
endif()

# 包含目录
include_directories(
    ${CMAKE_CURRENT_SOURCE_DIR}/include
    ${OPENSSL_INCLUDE_DIR}
    ${JsonCpp_INCLUDE_DIRS}
)

# 源文件定义
set(HEADER_FILES
    include/volunteer_matcher.h
    include/admission_predictor.h
    include/university_filter.h
    include/risk_assessor.h
)

set(SOURCE_FILES
    src/volunteer_matcher.cpp
    src/admission_predictor.cpp
    src/university_filter.cpp
    src/risk_assessor.cpp
    src/c_interface.cpp
)

# 创建一个接口库用于编译定义
add_library(volunteer_matcher_compile_defs INTERFACE)
target_compile_definitions(volunteer_matcher_compile_defs INTERFACE
    VOLUNTEER_MATCHER_VERSION_MAJOR=${PROJECT_VERSION_MAJOR}
    VOLUNTEER_MATCHER_VERSION_MINOR=${PROJECT_VERSION_MINOR}
    VOLUNTEER_MATCHER_VERSION_PATCH=${PROJECT_VERSION_PATCH}
)

# 静态库目标
add_library(volunteer_matcher_static STATIC
    ${HEADER_FILES}
    ${SOURCE_FILES}
)

# 动态库目标
add_library(volunteer_matcher_shared SHARED
    ${HEADER_FILES}
    ${SOURCE_FILES}
)

# 设置库别名
add_library(VolunteerMatcher::static ALIAS volunteer_matcher_static)
add_library(VolunteerMatcher::shared ALIAS volunteer_matcher_shared)

# 库属性设置
set_target_properties(volunteer_matcher_static PROPERTIES
    OUTPUT_NAME volunteer_matcher
    CLEAN_DIRECT_OUTPUT 1
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
    POSITION_INDEPENDENT_CODE ON
)

set_target_properties(volunteer_matcher_shared PROPERTIES
    OUTPUT_NAME volunteer_matcher
    CLEAN_DIRECT_OUTPUT 1
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
    POSITION_INDEPENDENT_CODE ON
)

# 链接库 - 静态库
target_link_libraries(volunteer_matcher_static
    PUBLIC
        volunteer_matcher_compile_defs
    PRIVATE
        OpenSSL::SSL
        OpenSSL::Crypto
        ${JsonCpp_LIBRARIES}
        ${PLATFORM_LIBS}
)

# 链接库 - 动态库
target_link_libraries(volunteer_matcher_shared
    PUBLIC
        volunteer_matcher_compile_defs
    PRIVATE
        OpenSSL::SSL
        OpenSSL::Crypto
        ${JsonCpp_LIBRARIES}
        ${PLATFORM_LIBS}
)

# XGBoost链接 (如果可用)
if(xgboost_FOUND)
    target_link_libraries(volunteer_matcher_static PRIVATE xgboost::xgboost)
    target_link_libraries(volunteer_matcher_shared PRIVATE xgboost::xgboost)
endif()

# OpenMP链接 (如果可用)
if(OpenMP_CXX_FOUND)
    target_link_libraries(volunteer_matcher_static PRIVATE OpenMP::OpenMP_CXX)
    target_link_libraries(volunteer_matcher_shared PRIVATE OpenMP::OpenMP_CXX)
endif()

# C接口导出(仅动态库)
if(WIN32)
    set_target_properties(volunteer_matcher_shared PROPERTIES
        WINDOWS_EXPORT_ALL_SYMBOLS ON
    )
endif()

# 编译定义
target_compile_definitions(volunteer_matcher_static
    PRIVATE
        VOLUNTEER_MATCHER_STATIC_LIB=1
    PUBLIC
        $<INSTALL_INTERFACE:VOLUNTEER_MATCHER_STATIC_LIB=1>
)

target_compile_definitions(volunteer_matcher_shared
    PRIVATE
        VOLUNTEER_MATCHER_SHARED_LIB=1
    PUBLIC
        VOLUNTEER_MATCHER_DLL=1
)

# 包含目录设置
target_include_directories(volunteer_matcher_static
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

target_include_directories(volunteer_matcher_shared
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
        $<INSTALL_INTERFACE:include>
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src
)

# 示例程序
option(BUILD_EXAMPLES "Build example programs" ON)
if(BUILD_EXAMPLES)
    # C++示例
    add_executable(volunteer_matcher_example
        examples/volunteer_matcher_example.cpp
    )
    
    target_link_libraries(volunteer_matcher_example
        PRIVATE
            volunteer_matcher_static
    )
    
    # C接口示例
    add_executable(c_interface_example
        examples/c_interface_example.c
    )
    
    target_link_libraries(c_interface_example
        PRIVATE
            volunteer_matcher_shared
    )
    
    set_target_properties(c_interface_example PROPERTIES
        C_STANDARD 99
        C_STANDARD_REQUIRED ON
    )
endif()

# 单元测试
if(BUILD_TESTS)
    enable_testing()
    
    # 主测试可执行文件
    add_executable(volunteer_matcher_tests
        tests/test_volunteer_matcher.cpp
        tests/test_admission_predictor.cpp
        tests/test_university_filter.cpp
        tests/test_risk_assessor.cpp
    )
    
    if(GTest_FOUND)
        target_link_libraries(volunteer_matcher_tests
            PRIVATE
                volunteer_matcher_static
                GTest::gtest
                GTest::gtest_main
                GTest::gmock
        )
    else()
        target_link_libraries(volunteer_matcher_tests
            PRIVATE
                volunteer_matcher_static
                gtest
                gtest_main
                gmock
        )
    endif()
    
    # 添加测试
    add_test(
        NAME VolunteerMatcherUnitTests
        COMMAND volunteer_matcher_tests
    )
    
    # 设置测试环境
    set_tests_properties(VolunteerMatcherUnitTests
        PROPERTIES
            TIMEOUT 600
            ENVIRONMENT "GTEST_OUTPUT=xml:${CMAKE_BINARY_DIR}/test_results.xml"
    )
    
    # 集成测试
    add_executable(integration_tests
        tests/integration_test.cpp
    )
    
    target_link_libraries(integration_tests
        PRIVATE
            volunteer_matcher_static
            GTest::gtest
            GTest::gtest_main
    )
    
    add_test(
        NAME VolunteerMatcherIntegrationTests
        COMMAND integration_tests
    )
endif()

# 性能基准测试
if(BUILD_BENCHMARKS AND benchmark_FOUND)
    add_executable(volunteer_matcher_benchmark
        benchmarks/volunteer_matcher_benchmark.cpp
        benchmarks/admission_predictor_benchmark.cpp
        benchmarks/university_filter_benchmark.cpp
    )
    
    target_link_libraries(volunteer_matcher_benchmark
        PRIVATE
            volunteer_matcher_static
            benchmark::benchmark
    )
    
    add_test(
        NAME VolunteerMatcherBenchmarks
        COMMAND volunteer_matcher_benchmark
    )
endif()

# Go绑定生成
option(BUILD_GO_BINDINGS "Generate Go language bindings" ON)
if(BUILD_GO_BINDINGS)
    # 创建Go模块文件
    configure_file(
        ${CMAKE_CURRENT_SOURCE_DIR}/bindings/go/volunteer_matcher.go.in
        ${CMAKE_CURRENT_BINARY_DIR}/bindings/go/volunteer_matcher.go
        @ONLY
    )
    
    # 生成CGO头文件
    add_custom_target(go_bindings ALL
        COMMAND ${CMAKE_COMMAND} -E copy_if_different
                ${CMAKE_CURRENT_SOURCE_DIR}/include/volunteer_matcher.h
                ${CMAKE_CURRENT_BINARY_DIR}/bindings/go/volunteer_matcher.h
        DEPENDS volunteer_matcher_shared
    )
endif()

# 安装配置
include(GNUInstallDirs)

# 安装头文件
install(
    FILES ${HEADER_FILES}
    DESTINATION ${CMAKE_INSTALL_INCLUDEDIR}/volunteer_matcher
    COMPONENT Development
)

# 安装库文件
install(
    TARGETS volunteer_matcher_static volunteer_matcher_shared volunteer_matcher_compile_defs
    EXPORT VolunteerMatcherTargets
    ARCHIVE DESTINATION ${CMAKE_INSTALL_LIBDIR}
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
    COMPONENT Runtime
)

# 安装示例程序
if(BUILD_EXAMPLES)
    install(
        TARGETS volunteer_matcher_example c_interface_example
        DESTINATION ${CMAKE_INSTALL_BINDIR}
        COMPONENT Examples
    )
endif()

# 生成和安装CMake配置文件
install(
    EXPORT VolunteerMatcherTargets
    FILE VolunteerMatcherTargets.cmake
    NAMESPACE VolunteerMatcher::
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/VolunteerMatcher
    COMPONENT Development
)

# 生成配置文件
include(CMakePackageConfigHelpers)

configure_package_config_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/cmake/VolunteerMatcherConfig.cmake.in
    ${CMAKE_CURRENT_BINARY_DIR}/VolunteerMatcherConfig.cmake
    INSTALL_DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/VolunteerMatcher
)

write_basic_package_version_file(
    ${CMAKE_CURRENT_BINARY_DIR}/VolunteerMatcherConfigVersion.cmake
    VERSION ${PROJECT_VERSION}
    COMPATIBILITY SameMajorVersion
)

install(
    FILES
        ${CMAKE_CURRENT_BINARY_DIR}/VolunteerMatcherConfig.cmake
        ${CMAKE_CURRENT_BINARY_DIR}/VolunteerMatcherConfigVersion.cmake
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/cmake/VolunteerMatcher
    COMPONENT Development
)

# pkg-config文件
configure_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/cmake/volunteer_matcher.pc.in
    ${CMAKE_CURRENT_BINARY_DIR}/volunteer_matcher.pc
    @ONLY
)

install(
    FILES ${CMAKE_CURRENT_BINARY_DIR}/volunteer_matcher.pc
    DESTINATION ${CMAKE_INSTALL_LIBDIR}/pkgconfig
    COMPONENT Development
)

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
find_package(Doxygen QUIET)
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

# 性能分析
add_custom_target(profile
    COMMAND valgrind --tool=callgrind --dump-instr=yes --collect-jumps=yes
            $<TARGET_FILE:volunteer_matcher_tests>
    DEPENDS volunteer_matcher_tests
    COMMENT "Running performance profiling"
)

# 内存检查
add_custom_target(memcheck
    COMMAND valgrind --tool=memcheck --leak-check=full --show-leak-kinds=all
            $<TARGET_FILE:volunteer_matcher_tests>
    DEPENDS volunteer_matcher_tests
    COMMENT "Running memory check"
)

# 清理目标
add_custom_target(distclean
    COMMAND ${CMAKE_COMMAND} -E remove_directory ${CMAKE_BINARY_DIR}/CMakeFiles
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_BINARY_DIR}/CMakeCache.txt
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_BINARY_DIR}/cmake_install.cmake
    COMMAND ${CMAKE_COMMAND} -E remove ${CMAKE_BINARY_DIR}/Makefile
    WORKING_DIRECTORY ${CMAKE_BINARY_DIR}
    COMMENT "Cleaning build directory"
)

# 信息输出
message(STATUS "")
message(STATUS "Volunteer Matcher Module Configuration Summary:")
message(STATUS "  Version:          ${PROJECT_VERSION}")
message(STATUS "  Build Type:       ${CMAKE_BUILD_TYPE}")
message(STATUS "  Platform:         ${CMAKE_SYSTEM_NAME}")
message(STATUS "  Architecture:     ${CMAKE_SYSTEM_PROCESSOR}")
message(STATUS "  Compiler:         ${CMAKE_CXX_COMPILER_ID} ${CMAKE_CXX_COMPILER_VERSION}")
message(STATUS "  C++ Standard:     ${CMAKE_CXX_STANDARD}")
message(STATUS "  OpenSSL Version:  ${OPENSSL_VERSION}")
message(STATUS "  Install Prefix:   ${CMAKE_INSTALL_PREFIX}")
message(STATUS "")
message(STATUS "Dependencies:")
message(STATUS "  OpenSSL:          ${OpenSSL_FOUND}")
message(STATUS "  JsonCpp:          ${JsonCpp_FOUND}")
message(STATUS "  XGBoost:          ${xgboost_FOUND}")
message(STATUS "  OpenMP:           ${OpenMP_CXX_FOUND}")
message(STATUS "")
message(STATUS "Build Options:")
message(STATUS "  Build Tests:      ${BUILD_TESTS}")
message(STATUS "  Build Examples:   ${BUILD_EXAMPLES}")
message(STATUS "  Build Benchmarks: ${BUILD_BENCHMARKS}")
message(STATUS "  Build Go Bindings:${BUILD_GO_BINDINGS}")
message(STATUS "")

# 编译后信息
add_custom_target(build-info ALL
    COMMAND ${CMAKE_COMMAND} -E echo "Volunteer Matcher Module build completed successfully!"
    COMMAND ${CMAKE_COMMAND} -E echo "Static Library: ${CMAKE_BINARY_DIR}/libvolunteer_matcher.a"
    COMMAND ${CMAKE_COMMAND} -E echo "Shared Library: ${CMAKE_BINARY_DIR}/libvolunteer_matcher.so"
    VERBATIM
>>>>>>> 0dd6b27ce36fbec25f47c1952ba01974d6d592bc
)