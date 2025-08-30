#!/bin/bash

# 高考志愿填报推荐服务 - 构建和测试脚本

set -e

echo "=== 高考志愿填报推荐服务 - 构建和测试 ==="

# 设置变量
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICE_NAME="recommendation-service"
VERSION=${VERSION:-latest}
BUILD_ENV=${BUILD_ENV:-dev}

echo "项目根目录: $PROJECT_ROOT"
echo "服务名称: $SERVICE_NAME"
echo "版本: $VERSION"
echo "构建环境: $BUILD_ENV"

# 函数：检查依赖
check_dependencies() {
    echo "检查构建依赖..."
    
    # 检查Go
    if ! command -v go &> /dev/null; then
        echo "错误: Go 未安装"
        exit 1
    fi
    echo "✓ Go $(go version | cut -d' ' -f3)"
    
    # 检查g++
    if ! command -v g++ &> /dev/null; then
        echo "错误: g++ 未安装"
        exit 1
    fi
    echo "✓ g++ $(g++ --version | head -1)"
    
    # 检查Docker（可选）
    if command -v docker &> /dev/null; then
        echo "✓ Docker $(docker --version | cut -d' ' -f3 | tr -d ',')"
    else
        echo "⚠ Docker 未安装（将跳过容器化构建）"
    fi
}

# 函数：构建C++模块
build_cpp_modules() {
    echo "构建C++模块..."
    
    CPP_MODULE_DIR="$PROJECT_ROOT/../../cpp-modules"
    
    if [ ! -d "$CPP_MODULE_DIR" ]; then
        echo "错误: C++模块目录不存在: $CPP_MODULE_DIR"
        exit 1
    fi
    
    # 构建志愿匹配器
    VOLUNTEER_MATCHER_DIR="$CPP_MODULE_DIR/volunteer-matcher"
    if [ -d "$VOLUNTEER_MATCHER_DIR" ]; then
        echo "构建志愿匹配器..."
        cd "$VOLUNTEER_MATCHER_DIR"
        
        mkdir -p build
        cd build
        
        cmake .. -DCMAKE_BUILD_TYPE=Release
        make -j$(nproc 2>/dev/null || echo 4)
        
        echo "✓ 志愿匹配器构建完成"
    else
        echo "⚠ 志愿匹配器目录不存在，跳过"
    fi
    
    # 构建AI推荐引擎
    AI_ENGINE_DIR="$CPP_MODULE_DIR/ai-recommendation-engine"
    if [ -d "$AI_ENGINE_DIR" ]; then
        echo "构建AI推荐引擎..."
        cd "$AI_ENGINE_DIR"
        
        mkdir -p build
        cd build
        
        cmake .. -DCMAKE_BUILD_TYPE=Release
        make -j$(nproc 2>/dev/null || echo 4)
        
        echo "✓ AI推荐引擎构建完成"
    else
        echo "⚠ AI推荐引擎目录不存在，跳过"
    fi
    
    # 构建混合推荐引擎
    HYBRID_ENGINE_DIR="$CPP_MODULE_DIR/hybrid-recommendation-engine"
    if [ -d "$HYBRID_ENGINE_DIR" ]; then
        echo "构建混合推荐引擎..."
        cd "$HYBRID_ENGINE_DIR"
        
        mkdir -p build
        cd build
        
        cmake .. -DCMAKE_BUILD_TYPE=Release
        make -j$(nproc 2>/dev/null || echo 4)
        
        echo "✓ 混合推荐引擎构建完成"
    else
        echo "⚠ 混合推荐引擎目录不存在，跳过"
    fi
    
    cd "$PROJECT_ROOT"
}

# 函数：构建Go服务
build_go_service() {
    echo "构建Go服务..."
    
    cd "$PROJECT_ROOT"
    
    # 更新依赖
    echo "更新Go模块依赖..."
    go mod tidy
    
    # 设置构建标志
    BUILD_FLAGS="-v"
    if [ "$BUILD_ENV" = "prod" ]; then
        BUILD_FLAGS="$BUILD_FLAGS -ldflags='-w -s'"
    fi
    
    # 构建服务
    echo "编译推荐服务..."
    go build $BUILD_FLAGS -o bin/${SERVICE_NAME} ./main.go
    
    if [ $? -eq 0 ]; then
        echo "✓ Go服务构建完成"
        echo "可执行文件: bin/${SERVICE_NAME}"
    else
        echo "❌ Go服务构建失败"
        exit 1
    fi
}

# 函数：运行测试
run_tests() {
    echo "运行测试..."
    
    cd "$PROJECT_ROOT"
    
    # 运行Go单元测试
    echo "运行Go单元测试..."
    if go test ./... -v; then
        echo "✓ Go单元测试通过"
    else
        echo "⚠ Go单元测试失败"
    fi
    
    # 运行go vet检查
    echo "运行代码静态分析..."
    if go vet ./...; then
        echo "✓ 静态分析通过"
    else
        echo "⚠ 静态分析发现问题"
    fi
    
    # 检查Go代码格式
    echo "检查代码格式..."
    UNFORMATTED=$(go fmt ./...)
    if [ -z "$UNFORMATTED" ]; then
        echo "✓ 代码格式正确"
    else
        echo "⚠ 以下文件格式需要调整:"
        echo "$UNFORMATTED"
    fi
}

# 函数：运行性能测试
run_performance_test() {
    echo "准备性能测试..."
    
    # 检查服务是否在运行
    if ! curl -s http://localhost:8083/health > /dev/null; then
        echo "启动推荐服务进行性能测试..."
        
        # 后台启动服务
        cd "$PROJECT_ROOT"
        ./bin/${SERVICE_NAME} &
        SERVICE_PID=$!
        
        # 等待服务启动
        echo "等待服务启动..."
        for i in {1..30}; do
            if curl -s http://localhost:8083/health > /dev/null; then
                echo "✓ 服务已启动"
                break
            fi
            sleep 1
        done
        
        if [ $i -eq 30 ]; then
            echo "❌ 服务启动超时"
            kill $SERVICE_PID 2>/dev/null || true
            exit 1
        fi
    else
        echo "✓ 服务已在运行"
        SERVICE_PID=""
    fi
    
    # 运行性能测试
    echo "运行性能测试..."
    cd "$PROJECT_ROOT/scripts"
    if go run performance_test.go; then
        echo "✓ 性能测试完成"
    else
        echo "⚠ 性能测试失败"
    fi
    
    # 清理
    if [ ! -z "$SERVICE_PID" ]; then
        echo "停止测试服务..."
        kill $SERVICE_PID 2>/dev/null || true
        wait $SERVICE_PID 2>/dev/null || true
    fi
}

# 函数：构建Docker镜像
build_docker_image() {
    if ! command -v docker &> /dev/null; then
        echo "跳过Docker构建（Docker未安装）"
        return
    fi
    
    echo "构建Docker镜像..."
    
    cd "$PROJECT_ROOT"
    
    # 检查Dockerfile
    if [ ! -f "Dockerfile" ]; then
        echo "⚠ Dockerfile不存在，跳过Docker构建"
        return
    fi
    
    # 构建镜像
    IMAGE_TAG="gaokao/${SERVICE_NAME}:${VERSION}"
    
    echo "构建镜像: $IMAGE_TAG"
    if docker build -t "$IMAGE_TAG" .; then
        echo "✓ Docker镜像构建完成"
        
        # 显示镜像信息
        docker images | grep "gaokao/${SERVICE_NAME}"
    else
        echo "❌ Docker镜像构建失败"
        exit 1
    fi
}

# 函数：清理
cleanup() {
    echo "清理构建缓存..."
    
    cd "$PROJECT_ROOT"
    
    # 清理Go缓存
    go clean -cache
    
    # 清理构建产物（可选）
    if [ "$1" = "full" ]; then
        rm -rf bin/
        rm -rf build/
        echo "✓ 完全清理完成"
    else
        echo "✓ 缓存清理完成"
    fi
}

# 主函数
main() {
    case "${1:-build}" in
        "check")
            check_dependencies
            ;;
        "cpp")
            check_dependencies
            build_cpp_modules
            ;;
        "go")
            check_dependencies
            build_go_service
            ;;
        "test")
            run_tests
            ;;
        "perf")
            run_performance_test
            ;;
        "docker")
            check_dependencies
            build_docker_image
            ;;
        "build")
            check_dependencies
            build_cpp_modules
            build_go_service
            run_tests
            ;;
        "all")
            check_dependencies
            build_cpp_modules
            build_go_service
            run_tests
            build_docker_image
            ;;
        "clean")
            cleanup "${2:-normal}"
            ;;
        *)
            echo "用法: $0 {check|cpp|go|test|perf|docker|build|all|clean}"
            echo ""
            echo "命令说明:"
            echo "  check  - 检查构建依赖"
            echo "  cpp    - 仅构建C++模块"
            echo "  go     - 仅构建Go服务"
            echo "  test   - 运行测试"
            echo "  perf   - 运行性能测试"
            echo "  docker - 构建Docker镜像"
            echo "  build  - 构建C++模块、Go服务并运行测试"
            echo "  all    - 完整构建（包括Docker镜像）"
            echo "  clean  - 清理构建缓存 (clean full: 完全清理)"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"