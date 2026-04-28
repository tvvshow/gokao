cmake_minimum_required(VERSION 3.15)
project(GaoKaoHubLicense VERSION 1.0.0 LANGUAGES CXX)

# 设置C++标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# 设置构建类型
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

# 编译选项
set(CMAKE_CXX_FLAGS "-Wall -Wextra")
set(CMAKE_CXX_FLAGS_DEBUG "-g -O0 -DDEBUG")
set(CMAKE_CXX_FLAGS_RELEASE "-O3 -DNDEBUG")

# 安全编译选项
if(CMAKE_CXX_COMPILER_ID STREQUAL "GNU" OR CMAKE_CXX_COMPILER_ID STREQUAL "Clang")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fstack-protector-strong -D_FORTIFY_SOURCE=2")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -fPIE -Wformat -Wformat-security")
    
    if(CMAKE_BUILD_TYPE STREQUAL "Release")
        set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -ffunction-sections -fdata-sections")
        set(CMAKE_EXE_LINKER_FLAGS "${CMAKE_EXE_LINKER_FLAGS} -Wl,--gc-sections")
    endif()
endif()

# 查找依赖包
find_package(PkgConfig REQUIRED)

# OpenSSL
find_package(OpenSSL REQUIRED)
if(OpenSSL_FOUND)
    message(STATUS "Found OpenSSL ${OPENSSL_VERSION}")
else()
    message(FATAL_ERROR "OpenSSL not found")
endif()

# SQLite3
pkg_check_modules(SQLITE3 REQUIRED sqlite3)

# JsonCpp
pkg_check_modules(JSONCPP REQUIRED jsoncpp)

# 包含目录
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/include)
include_directories(${OPENSSL_INCLUDE_DIR})
include_directories(${SQLITE3_INCLUDE_DIRS})
include_directories(${JSONCPP_INCLUDE_DIRS})

# 源文件
set(SOURCES
    src/license_manager.cpp
    src/crypto_utils.cpp
)

# 头文件
set(HEADERS
    include/license_manager.h
    include/crypto_utils.h
)

if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/src/c_interface.cpp")
    list(APPEND SOURCES src/c_interface.cpp)
endif()

if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/include/c_interface.h")
    list(APPEND HEADERS include/c_interface.h)
endif()

# 静态库
add_library(gaokaohub_license_static STATIC ${SOURCES} ${HEADERS})
target_link_libraries(gaokaohub_license_static 
    ${OPENSSL_LIBRARIES}
    ${SQLITE3_LIBRARIES}
    ${JSONCPP_LIBRARIES}
)

# 动态库
add_library(gaokaohub_license_shared SHARED ${SOURCES} ${HEADERS})
target_link_libraries(gaokaohub_license_shared 
    ${OPENSSL_LIBRARIES}
    ${SQLITE3_LIBRARIES}
    ${JSONCPP_LIBRARIES}
)

# 设置库属性
set_target_properties(gaokaohub_license_static PROPERTIES
    OUTPUT_NAME gaokaohub_license
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
)

set_target_properties(gaokaohub_license_shared PROPERTIES
    OUTPUT_NAME gaokaohub_license
    VERSION ${PROJECT_VERSION}
    SOVERSION ${PROJECT_VERSION_MAJOR}
    POSITION_INDEPENDENT_CODE ON
)

# 编译定义
target_compile_definitions(gaokaohub_license_static PRIVATE
    GAOKAOHUB_LICENSE_STATIC
    $<$<CONFIG:Debug>:GAOKAOHUB_DEBUG>
)

target_compile_definitions(gaokaohub_license_shared PRIVATE
    GAOKAOHUB_LICENSE_SHARED
    $<$<CONFIG:Debug>:GAOKAOHUB_DEBUG>
)

# 测试可执行文件
if(BUILD_TESTING
   AND EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/tests/test_license_manager.cpp"
   AND EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/tests/test_crypto_utils.cpp")
    enable_testing()
    
    # 查找测试框架
    find_package(GTest)
    if(GTest_FOUND)
        add_executable(license_test
            tests/test_license_manager.cpp
            tests/test_crypto_utils.cpp
        )
        
        target_link_libraries(license_test
            gaokaohub_license_static
            GTest::gtest
            GTest::gtest_main
            ${OPENSSL_LIBRARIES}
            ${SQLITE3_LIBRARIES}
            ${JSONCPP_LIBRARIES}
        )
        
        add_test(NAME LicenseTest COMMAND license_test)
    endif()
endif()

# 示例程序
if(EXISTS "${CMAKE_CURRENT_SOURCE_DIR}/examples/license_example.cpp")
    add_executable(license_example examples/license_example.cpp)
    target_link_libraries(license_example 
        gaokaohub_license_static
        ${OPENSSL_LIBRARIES}
        ${SQLITE3_LIBRARIES}
        ${JSONCPP_LIBRARIES}
    )
endif()

# 安装配置
install(TARGETS gaokaohub_license_static gaokaohub_license_shared
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
)

install(FILES ${HEADERS}
    DESTINATION include/gaokaohub/license
)

# 生成pkg-config文件
configure_file(
    ${CMAKE_CURRENT_SOURCE_DIR}/gaokaohub-license.pc.in
    ${CMAKE_CURRENT_BINARY_DIR}/gaokaohub-license.pc
    @ONLY
)

install(FILES ${CMAKE_CURRENT_BINARY_DIR}/gaokaohub-license.pc
    DESTINATION lib/pkgconfig
)

# CPack配置
set(CPACK_PACKAGE_NAME "gaokaohub-license")
set(CPACK_PACKAGE_VENDOR "GaoKaoHub")
set(CPACK_PACKAGE_DESCRIPTION_SUMMARY "GaoKaoHub License Management Library")
set(CPACK_PACKAGE_VERSION ${PROJECT_VERSION})
set(CPACK_PACKAGE_VERSION_MAJOR ${PROJECT_VERSION_MAJOR})
set(CPACK_PACKAGE_VERSION_MINOR ${PROJECT_VERSION_MINOR})
set(CPACK_PACKAGE_VERSION_PATCH ${PROJECT_VERSION_PATCH})

include(CPack)

# 打印构建信息
message(STATUS "")
message(STATUS "================ Build Configuration ================")
message(STATUS "Build Type: ${CMAKE_BUILD_TYPE}")
message(STATUS "C++ Standard: ${CMAKE_CXX_STANDARD}")
message(STATUS "Compiler: ${CMAKE_CXX_COMPILER_ID} ${CMAKE_CXX_COMPILER_VERSION}")
message(STATUS "OpenSSL Version: ${OPENSSL_VERSION}")
message(STATUS "Install Prefix: ${CMAKE_INSTALL_PREFIX}")
message(STATUS "====================================================")
message(STATUS "")
