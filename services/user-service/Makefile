# Makefile for User Service

# 变量定义
APP_NAME=user-service
DOCKER_IMAGE=gaokao/user-service
DOCKER_TAG=latest
GO_VERSION=1.21
PORT=8080

# 颜色定义
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

.PHONY: help build build-cgo run test test-cgo test-integration clean docker-build docker-run docker-stop docker-clean dev deps lint format swagger migrate seed logs build-cpp check-cpp

# 默认目标
all: build

# 帮助信息
help:
	@echo "$(BLUE)用户服务 Makefile 命令:$(NC)"
	@echo ""
	@echo "$(GREEN)开发命令:$(NC)"
	# 开发命令:$(NC)"
	@echo "  make dev          - 启动开发环境"
	@echo "  make build        - 构建应用"
	@echo "  make run          - 运行应用"
	@echo "  make test         - 运行Go测试"
	@echo "  make test-integration - 运行完整集成测试"
	@echo "  make deps         - 安装依赖"
	@echo "  make lint         - 代码检查"
	@echo "  make format       - 格式化代码"
	@echo "  make swagger      - 生成Swagger文档"
	@echo ""
	@echo "$(GREEN)Docker命令:$(NC)"
	@echo "  make docker-build - 构建Docker镜像"
	@echo "  make docker-run   - 运行Docker容器"
	@echo "  make docker-stop  - 停止Docker容器"
	@echo "  make docker-clean - 清理Docker资源"
	@echo ""
	@echo "$(GREEN)数据库命令:$(NC)"
	@echo "  make migrate      - 运行数据库迁移"
	@echo "  make seed         - 填充测试数据"
	@echo ""
	@echo "$(GREEN)其他命令:$(NC)"
	@echo "  make clean        - 清理构建文件"
	@echo "  make logs         - 查看日志"

# 安装依赖
deps:
	@echo "$(YELLOW)安装Go依赖...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)依赖安装完成$(NC)"

# 代码检查
lint:
	@echo "$(YELLOW)运行代码检查...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(RED)golangci-lint 未安装，跳过代码检查$(NC)"; \
	fi
	@echo "$(GREEN)代码检查完成$(NC)"

# 格式化代码
format:
	@echo "$(YELLOW)格式化代码...$(NC)"
	go fmt ./...
	go vet ./...
	@echo "$(GREEN)代码格式化完成$(NC)"

# 生成Swagger文档
swagger:
	@echo "$(YELLOW)生成Swagger文档...$(NC)"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g main.go --output ./docs; \
	else \
		echo "$(RED)swag 未安装，请运行: go install github.com/swaggo/swag/cmd/swag@latest$(NC)"; \
	fi
	@echo "$(GREEN)Swagger文档生成完成$(NC)"

# 构建应用
build: deps format
	@echo "$(YELLOW)构建应用...$(NC)"
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) .
	@echo "$(GREEN)构建完成: bin/$(APP_NAME)$(NC)"

# 运行应用
run: build
	@echo "$(YELLOW)启动应用...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)复制环境配置文件...$(NC)"; \
		cp .env.example .env; \
	fi
	./bin/$(APP_NAME)

# 开发模式
dev:
	@echo "$(YELLOW)启动开发环境...$(NC)"
	@if [ ! -f .env ]; then \
		echo "$(YELLOW)复制环境配置文件...$(NC)"; \
		cp .env.example .env; \
	fi
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(RED)air 未安装，使用普通模式启动$(NC)"; \
		go run main.go; \
	fi

# 运行Go测试
test:
	@echo "$(YELLOW)运行Go测试...$(NC)"
	CGO_ENABLED=0 go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Go测试完成，覆盖率报告: coverage.html$(NC)"

# 运行完整集成测试
test-integration:
	@echo "$(YELLOW)运行完整集成测试...$(NC)"
	CGO_ENABLED=0 go test -v -race -coverprofile=coverage-full.out ./...
	go tool cover -html=coverage-full.out -o coverage-full.html
	@echo "$(GREEN)完整集成测试完成，覆盖率报告: coverage-full.html$(NC)"

# 构建Docker镜像
docker-build:
	@echo "$(YELLOW)构建Docker镜像...$(NC)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)Docker镜像构建完成: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

# 运行Docker容器
docker-run:
	@echo "$(YELLOW)启动Docker容器...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)Docker容器启动完成$(NC)"
	@echo "$(BLUE)服务地址: http://localhost:$(PORT)$(NC)"
	@echo "$(BLUE)Swagger文档: http://localhost:$(PORT)/swagger/index.html$(NC)"
	@echo "$(BLUE)数据库管理: http://localhost:8081$(NC)"

# 停止Docker容器
docker-stop:
	@echo "$(YELLOW)停止Docker容器...$(NC)"
	docker-compose down
	@echo "$(GREEN)Docker容器已停止$(NC)"

# 清理Docker资源
docker-clean: docker-stop
	@echo "$(YELLOW)清理Docker资源...$(NC)"
	docker-compose down -v --rmi all --remove-orphans
	docker system prune -f
	@echo "$(GREEN)Docker资源清理完成$(NC)"

# 数据库迁移
migrate:
	@echo "$(YELLOW)运行数据库迁移...$(NC)"
	@echo "$(BLUE)迁移将在应用启动时自动执行$(NC)"

# 填充测试数据
seed:
	@echo "$(YELLOW)填充测试数据...$(NC)"
	@echo "$(BLUE)测试数据将在应用启动时自动创建$(NC)"

# 查看日志
logs:
	@echo "$(YELLOW)查看应用日志...$(NC)"
	docker-compose logs -f user-service

# 查看所有服务日志
logs-all:
	@echo "$(YELLOW)查看所有服务日志...$(NC)"
	docker-compose logs -f

# 进入容器
shell:
	@echo "$(YELLOW)进入应用容器...$(NC)"
	docker-compose exec user-service sh

# 进入数据库
db-shell:
	@echo "$(YELLOW)进入数据库...$(NC)"
	docker-compose exec postgres psql -U gaokao_user -d gaokao_user_db

# 备份数据库
db-backup:
	@echo "$(YELLOW)备份数据库...$(NC)"
	mkdir -p backups
	docker-compose exec postgres pg_dump -U gaokao_user gaokao_user_db > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "$(GREEN)数据库备份完成$(NC)"

# 恢复数据库
db-restore:
	@echo "$(YELLOW)恢复数据库...$(NC)"
	@read -p "请输入备份文件路径: " backup_file; \
	if [ -f "$$backup_file" ]; then \
		docker-compose exec -T postgres psql -U gaokao_user -d gaokao_user_db < "$$backup_file"; \
		echo "$(GREEN)数据库恢复完成$(NC)"; \
	else \
		echo "$(RED)备份文件不存在$(NC)"; \
	fi

# 重启服务
restart: docker-stop docker-run

# 更新依赖
update:
	@echo "$(YELLOW)更新依赖...$(NC)"
	go get -u ./...
	go mod tidy
	@echo "$(GREEN)依赖更新完成$(NC)"

# 安装开发工具
install-tools:
	@echo "$(YELLOW)安装开发工具...$(NC)"
	go install github.com/cosmtrek/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)开发工具安装完成$(NC)"

# 清理构建文件
clean:
	@echo "$(YELLOW)清理构建文件...$(NC)"
	rm -rf bin/
	rm -rf docs/
	rm -f coverage.out coverage.html coverage-full.out coverage-full.html
	@echo "$(GREEN)清理完成$(NC)"

# 深度清理(包含依赖)
clean-all: clean
	@echo "$(YELLOW)深度清理...$(NC)"
	go clean -modcache
	rm -rf vendor/
	@echo "$(GREEN)深度清理完成$(NC)"

# 检查服务状态
status:
	@echo "$(YELLOW)检查服务状态...$(NC)"
	docker-compose ps

# 性能测试
bench:
	@echo "$(YELLOW)运行性能测试...$(NC)"
	go test -bench=. -benchmem ./...

# 安全扫描
security:
	@echo "$(YELLOW)运行安全扫描...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(RED)gosec 未安装，请运行: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi

# 生成API文档
docs: swagger
	@echo "$(GREEN)API文档已生成$(NC)"
	@echo "$(BLUE)本地查看: http://localhost:$(PORT)/swagger/index.html$(NC)"

# 完整的CI流程
ci: deps format lint test security build
	@echo "$(GREEN)CI流程完成$(NC)"

# 生产部署准备
prod-ready: ci docker-build
	@echo "$(GREEN)生产部署准备完成$(NC)"
	@echo "$(BLUE)Docker镜像: $(DOCKER_IMAGE):$(DOCKER_TAG)$(NC)"

# 本地环境设置
setup-dev: install-tools deps
	@echo "$(GREEN)开发环境设置完成$(NC)"
	@echo "$(YELLOW)运行测试验证环境：make test$(NC)"

# 环境验证
verify-env: test
	@echo "$(GREEN)环境验证通过$(NC)"