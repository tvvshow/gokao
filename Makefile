# 高考志愿填报系统 Makefile
# 支持Windows和Linux环境的构建管理

# 检测操作系统
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    EXE_EXT := .exe
    PATH_SEP := \\
    RM := del /Q
    MKDIR := mkdir
else
    DETECTED_OS := $(shell uname -s)
    EXE_EXT :=
    PATH_SEP := /
    RM := rm -f
    MKDIR := mkdir -p
endif

# 项目配置
PROJECT_NAME := gaokao-system
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 目录配置
BIN_DIR := bin
DIST_DIR := dist
COVERAGE_DIR := coverage

# Go配置
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 1

# 构建标志
LDFLAGS := -X main.Version=$(VERSION) \
           -X main.BuildTime=$(BUILD_TIME) \
           -X main.GitCommit=$(GIT_COMMIT)

ifeq ($(RELEASE),1)
    LDFLAGS += -s -w
    CGO_ENABLED := 0
endif

# 服务列表
SERVICES := api-gateway user-service data-service payment-service recommendation-service

# 默认目标
.PHONY: all
all: clean deps build test

# 帮助信息
.PHONY: help
help:
	@echo "高考志愿填报系统构建工具"
	@echo ""
	@echo "可用目标:"
	@echo "  all          - 完整构建流程"
	@echo "  build        - 构建所有组件"
	@echo "  build-go     - 构建Go服务"
	@echo "  build-frontend - 构建前端应用"
	@echo "  test         - 运行所有测试"
	@echo "  clean        - 清理构建产物"
	@echo "  deps         - 安装依赖"
	@echo "  dev          - 启动开发环境"
	@echo "  docker       - 构建Docker镜像"

# 清理
.PHONY: clean
clean:
	@echo "🧹 清理构建产物..."
	@$(RM) -r $(BIN_DIR) 2>/dev/null || true
	@$(RM) -r $(DIST_DIR) 2>/dev/null || true
	@$(RM) -r frontend$(PATH_SEP)dist 2>/dev/null || true
	@echo "✅ 清理完成"

# 安装依赖
.PHONY: deps
deps: deps-go deps-frontend

.PHONY: deps-go
deps-go:
	@echo "📦 安装Go依赖..."
	@go mod download
	@go mod tidy
	@echo "✅ Go依赖安装完成"

.PHONY: deps-frontend
deps-frontend:
	@echo "📦 安装前端依赖..."
	@cd frontend && npm ci
	@echo "✅ 前端依赖安装完成"

# 构建
.PHONY: build
build: build-go build-frontend

.PHONY: build-go
build-go:
	@echo "🔨 构建Go服务..."
	@$(MKDIR) $(BIN_DIR) 2>/dev/null || true
	@for service in $(SERVICES); do \
		echo "  构建 $$service..."; \
		cd services/$$service && \
		CGO_ENABLED=$(CGO_ENABLED) go build \
			-ldflags "$(LDFLAGS)" \
			-o ../../$(BIN_DIR)/$$service$(EXE_EXT) \
			. && \
		cd ../..; \
	done
	@echo "✅ Go服务构建完成"

.PHONY: build-frontend
build-frontend:
	@echo "🎨 构建前端应用..."
	@cd frontend && npm run build
	@echo "✅ 前端构建完成"

# 测试
.PHONY: test
test: test-go test-frontend

.PHONY: test-go
test-go:
	@echo "🧪 运行Go测试..."
	@GO_ENV=test go test -v -race -coverprofile=coverage.out ./...
	@echo "✅ Go测试完成"

.PHONY: test-frontend
test-frontend:
	@echo "🧪 运行前端测试..."
	@cd frontend && npm run test:unit
	@echo "✅ 前端测试完成"

# 开发模式
.PHONY: dev
dev:
	@echo "🔧 启动开发模式..."
	@$(MAKE) build-go
	@echo "✅ 开发环境准备完成"

# Docker构建
.PHONY: docker
docker:
	@echo "🐳 构建Docker镜像..."
	@for service in $(SERVICES); do \
		echo "  构建 $$service 镜像..."; \
		docker build -t $(PROJECT_NAME)/$$service:$(VERSION) \
			services/$$service; \
	done
	@docker build -t $(PROJECT_NAME)/frontend:$(VERSION) frontend
	@echo "✅ Docker镜像构建完成"

# 版本信息
.PHONY: version
version:
	@echo "项目: $(PROJECT_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo "操作系统: $(DETECTED_OS)"