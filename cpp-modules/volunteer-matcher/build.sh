#!/bin/bash

# 高考志愿填报系统 - C++志愿匹配引擎构建脚本
# 作者: 高考志愿填报系统开发团队
# 版本: 1.0.0
# 日期: 2025-01-18

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "Command '$1' not found. Please install $1 first."
        exit 1
    fi
}

# 检查依赖
check_dependencies() {
    log_info "Checking dependencies..."
    
    check_command "cmake"
    check_command "make"
    check_command "pkg-config"
    
    # 检查编译器
    if command -v g++ &> /dev/null; then
        log_info "Found g++ compiler"
    elif command -v clang++ &> /dev/null; then
        log_info "Found clang++ compiler"
    else
        log_error "No C++ compiler found. Please install g++ or clang++."
        exit 1
    fi
    
    # 检查JsonCpp
    if pkg-config --exists jsoncpp; then
        log_info "Found JsonCpp: $(pkg-config --modversion jsoncpp)"
    else
        log_warn "JsonCpp not found via pkg-config. Will try to find manually."
    fi
    
    # 检查OpenSSL
    if pkg-config --exists openssl; then
        log_info "Found OpenSSL: $(pkg-config --modversion openssl)"
    else
        log_warn "OpenSSL not found via pkg-config."
    fi
    
    log_success "Dependencies check completed"
}

# 清理构建目录
clean_build() {
    log_info "Cleaning build directory..."
    if [ -d "build" ]; then
        rm -rf build
    fi
    mkdir -p build
    log_success "Build directory cleaned"
}

# CMake配置
configure_cmake() {
    log_info "Configuring CMake..."
    cd build
    
    # 检测构建类型
    BUILD_TYPE=${1:-Release}
    log_info "Build type: $BUILD_TYPE"
    
    # CMake选项
    CMAKE_OPTIONS=(
        "-DCMAKE_BUILD_TYPE=$BUILD_TYPE"
        "-DBUILD_TESTS=ON"
        "-DBUILD_EXAMPLES=ON"
        "-DBUILD_GO_BINDINGS=ON"
    )
    
    # 添加额外选项
    if command -v ninja &> /dev/null; then
        CMAKE_OPTIONS+=("-GNinja")
        log_info "Using Ninja build system"
    fi
    
    # 运行CMake
    if cmake "${CMAKE_OPTIONS[@]}" ..; then
        log_success "CMake configuration successful"
    else
        log_error "CMake configuration failed"
        exit 1
    fi
    
    cd ..
}

# 编译项目
build_project() {
    log_info "Building project..."
    cd build
    
    # 检测CPU核心数
    if command -v nproc &> /dev/null; then
        CORES=$(nproc)
    elif command -v sysctl &> /dev/null; then
        CORES=$(sysctl -n hw.ncpu)
    else
        CORES=4
    fi
    
    log_info "Using $CORES parallel jobs"
    
    # 开始编译
    if command -v ninja &> /dev/null && [ -f "build.ninja" ]; then
        ninja -j$CORES
    else
        make -j$CORES
    fi
    
    if [ $? -eq 0 ]; then
        log_success "Build completed successfully"
    else
        log_error "Build failed"
        exit 1
    fi
    
    cd ..
}

# 运行测试
run_tests() {
    log_info "Running tests..."
    cd build
    
    if [ -f "volunteer_matcher_tests" ]; then
        ./volunteer_matcher_tests
        if [ $? -eq 0 ]; then
            log_success "All tests passed"
        else
            log_error "Some tests failed"
            return 1
        fi
    else
        log_warn "Test executable not found"
    fi
    
    cd ..
}

# 安装库文件
install_library() {
    log_info "Installing library..."
    cd build
    
    if [ "$EUID" -ne 0 ]; then
        log_warn "Not running as root. Using local install prefix."
        make install DESTDIR=./install_local
    else
        make install
    fi
    
    log_success "Library installed"
    cd ..
}

# 生成Go绑定
generate_go_bindings() {
    log_info "Generating Go bindings..."
    cd build
    
    if [ -d "bindings/go" ]; then
        log_info "Go bindings generated in build/bindings/go/"
        ls -la bindings/go/
    else
        log_warn "Go bindings not found"
    fi
    
    cd ..
}

# 显示构建信息
show_build_info() {
    log_info "Build Summary:"
    echo "======================================"
    
    if [ -d "build" ]; then
        cd build
        
        if [ -f "libvolunteer_matcher.a" ]; then
            SIZE=$(du -h libvolunteer_matcher.a | cut -f1)
            log_info "Static library: libvolunteer_matcher.a ($SIZE)"
        fi
        
        if [ -f "libvolunteer_matcher.so" ] || [ -f "libvolunteer_matcher.dylib" ]; then
            if [ -f "libvolunteer_matcher.so" ]; then
                SIZE=$(du -h libvolunteer_matcher.so | cut -f1)
                log_info "Shared library: libvolunteer_matcher.so ($SIZE)"
            else
                SIZE=$(du -h libvolunteer_matcher.dylib | cut -f1)
                log_info "Shared library: libvolunteer_matcher.dylib ($SIZE)"
            fi
        fi
        
        if [ -f "volunteer_matcher_example" ]; then
            log_info "Example executable: volunteer_matcher_example"
        fi
        
        if [ -f "volunteer_matcher_tests" ]; then
            log_info "Test executable: volunteer_matcher_tests"
        fi
        
        cd ..
    fi
    
    echo "======================================"
    log_success "Build completed successfully!"
}

# 主函数
main() {
    log_info "Starting C++ Volunteer Matcher build process..."
    
    # 解析命令行参数
    BUILD_TYPE="Release"
    RUN_TESTS=false
    INSTALL=false
    CLEAN=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --debug)
                BUILD_TYPE="Debug"
                shift
                ;;
            --test)
                RUN_TESTS=true
                shift
                ;;
            --install)
                INSTALL=true
                shift
                ;;
            --clean)
                CLEAN=true
                shift
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo "Options:"
                echo "  --debug    Build in debug mode"
                echo "  --test     Run tests after build"
                echo "  --install  Install library after build"
                echo "  --clean    Clean build directory before build"
                echo "  --help     Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # 执行构建步骤
    check_dependencies
    
    if [ "$CLEAN" = true ]; then
        clean_build
    fi
    
    if [ ! -d "build" ]; then
        mkdir -p build
    fi
    
    configure_cmake "$BUILD_TYPE"
    build_project
    generate_go_bindings
    
    if [ "$RUN_TESTS" = true ]; then
        run_tests
    fi
    
    if [ "$INSTALL" = true ]; then
        install_library
    fi
    
    show_build_info
    
    log_success "Build process completed!"
}

# 错误处理
trap 'log_error "Build process interrupted"; exit 1' INT TERM

# 运行主函数
main "$@"