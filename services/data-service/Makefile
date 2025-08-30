# Makefile for Data Service

# 变量定义
APP_NAME=data-service
DOCKER_IMAGE=gaokao/data-service
VERSION=1.0.0
GO_VERSION=1.23

# 颜色定义
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build run test clean docker-build docker-run deps fmt lint swagger

help: ## 显示帮助信息
	@echo "$(GREEN)数据服务 Makefile 帮助$(NC)"
	@echo ""
	@echo "$(YELLOW)可用命令:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(BLUE)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## 构建应用
	@echo "$(GREEN)构建应用...$(NC)"
	go build -o bin/$(APP_NAME) .
	@echo "$(GREEN)构建完成!$(NC)"

run: ## 运行应用
	@echo "$(GREEN)启动数据服务...$(NC)"
	go run main.go

test: ## 运行测试
	@echo "$(GREEN)运行测试...$(NC)"
	go test -v ./tests/...
	@echo "$(GREEN)测试完成!$(NC)"

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "$(GREEN)运行测试覆盖率检查...$(NC)"
	go test -v -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)覆盖率报告已生成: coverage.html$(NC)"

benchmark: ## 运行性能测试
	@echo "$(GREEN)运行性能测试...$(NC)"
	go test -bench=. -benchmem ./tests/...

clean: ## 清理构建文件
	@echo "$(GREEN)清理构建文件...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean
	@echo "$(GREEN)清理完成!$(NC)"

deps: ## 安装依赖
	@echo "$(GREEN)安装依赖...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)依赖安装完成!$(NC)"

fmt: ## 格式化代码
	@echo "$(GREEN)格式化代码...$(NC)"
	go fmt ./...
	@echo "$(GREEN)代码格式化完成!$(NC)"

lint: ## 代码静态检查
	@echo "$(GREEN)运行代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint 未安装，跳过代码检查$(NC)"; \
		echo "$(YELLOW)请运行: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

swagger: ## 生成Swagger文档
	@echo "$(GREEN)生成Swagger文档...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g main.go -o docs/; \
		echo "$(GREEN)Swagger文档生成完成!$(NC)"; \
	else \
		echo "$(YELLOW)swag 未安装，请运行: go install github.com/swaggo/swag/cmd/swag@latest$(NC)"; \
	fi

docker-build: ## 构建Docker镜像
	@echo "$(GREEN)构建Docker镜像...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker镜像构建完成!$(NC)"

docker-run: ## 运行Docker容器
	@echo "$(GREEN)启动Docker容器...$(NC)"
	docker run -d \
		--name $(APP_NAME) \
		-p 8082:8082 \
		-e GIN_MODE=release \
		$(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker容器启动完成!$(NC)"

docker-stop: ## 停止Docker容器
	@echo "$(GREEN)停止Docker容器...$(NC)"
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true
	@echo "$(GREEN)Docker容器已停止!$(NC)"

docker-logs: ## 查看Docker容器日志
	docker logs -f $(APP_NAME)

docker-push: ## 推送Docker镜像
	@echo "$(GREEN)推送Docker镜像...$(NC)"
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest
	@echo "$(GREEN)Docker镜像推送完成!$(NC)"

setup-dev: deps swagger ## 设置开发环境
	@echo "$(GREEN)开发环境设置完成!$(NC)"
	@echo "$(YELLOW)请复制 .env.example 到 .env 并配置相应参数$(NC)"

install-tools: ## 安装开发工具
	@echo "$(GREEN)安装开发工具...$(NC)"
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)开发工具安装完成!$(NC)"

check: fmt lint test ## 运行所有检查
	@echo "$(GREEN)所有检查完成!$(NC)"

release: clean deps swagger test docker-build ## 构建发布版本
	@echo "$(GREEN)发布版本构建完成!$(NC)"

db-migrate: ## 数据库迁移
	@echo "$(GREEN)运行数据库迁移...$(NC)"
	go run main.go --migrate
	@echo "$(GREEN)数据库迁移完成!$(NC)"

dev: ## 开发模式运行（热重载）
	@echo "$(GREEN)开发模式启动...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(YELLOW)air 未安装，使用普通模式启动$(NC)"; \
		echo "$(YELLOW)安装hot reload: go install github.com/cosmtrek/air@latest$(NC)"; \
		go run main.go; \
	fi

status: ## 显示应用状态
	@echo "$(GREEN)应用状态:$(NC)"
	@if docker ps | grep -q $(APP_NAME); then \
		echo "$(GREEN)✓ Docker容器运行中$(NC)"; \
		docker ps | grep $(APP_NAME); \
	else \
		echo "$(RED)✗ Docker容器未运行$(NC)"; \
	fi

info: ## 显示项目信息
	@echo "$(GREEN)项目信息:$(NC)"
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "Go版本: $(GO_VERSION)"
	@echo "Docker镜像: $(DOCKER_IMAGE)"

# 默认目标
.DEFAULT_GOAL := help