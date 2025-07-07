# Go-Kits 项目 Makefile
# 提供便捷的依赖管理和项目维护命令

.PHONY: help update sync verify test status clean fmt lint deps-update go-version-update

# 默认目标
.DEFAULT_GOAL := help

# 项目配置
PROJECT_NAME := go-kits
GO_VERSION := 1.24.3

# 所有模块目录
MODULES := . core logger config redis logger_v2 excel message storage influxdb etcd prometheus mqtt gout sqldb

help: ## 显示帮助信息
	@echo "Go-Kits 项目管理命令"
	@echo "===================="
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

update: ## 完整更新所有模块依赖
	@echo "🚀 开始更新所有模块依赖..."
	@chmod +x scripts/update-deps.sh
	@./scripts/update-deps.sh update

sync: ## 同步工作区
	@echo "🔄 同步工作区..."
	@go work sync

verify: ## 验证所有模块可以构建
	@echo "🔍 验证构建..."
	@go build ./...

test: ## 运行所有模块的测试
	@echo "🧪 运行测试..."
	@chmod +x scripts/update-deps.sh
	@./scripts/update-deps.sh test

status: ## 显示当前依赖状态
	@echo "📊 显示依赖状态..."
	@chmod +x scripts/update-deps.sh
	@./scripts/update-deps.sh status

clean: ## 清理构建缓存和依赖
	@echo "🧹 清理项目..."
	@go clean -cache
	@go clean -modcache -x
	@for dir in $(MODULES); do \
		if [ "$$dir" = "." ]; then \
			echo "  - 清理根模块"; \
		else \
			echo "  - 清理模块: $$dir"; \
			(cd $$dir && go clean); \
		fi; \
	done

fmt: ## 格式化所有 Go 代码
	@echo "📝 格式化代码..."
	@for dir in $(MODULES); do \
		if [ "$$dir" = "." ]; then \
			echo "  - 格式化根模块"; \
			go fmt ./...; \
		else \
			echo "  - 格式化模块: $$dir"; \
			(cd $$dir && go fmt ./...); \
		fi; \
	done

lint: ## 运行代码检查
	@echo "🔍 运行代码检查..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "安装 golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}
	@for dir in $(MODULES); do \
		if [ "$$dir" = "." ]; then \
			echo "  - 检查根模块"; \
			golangci-lint run ./...; \
		else \
			echo "  - 检查模块: $$dir"; \
			(cd $$dir && golangci-lint run ./...); \
		fi; \
	done

deps-update: ## 更新所有外部依赖到最新版本
	@echo "📦 更新外部依赖..."
	@for dir in $(MODULES); do \
		if [ "$$dir" = "." ]; then \
			echo "  - 更新根模块依赖"; \
			go get -u ./...; \
			go mod tidy; \
		else \
			echo "  - 更新模块依赖: $$dir"; \
			(cd $$dir && go get -u ./... && go mod tidy); \
		fi; \
	done
	@$(MAKE) sync

go-version-update: ## 更新所有模块的 Go 版本
	@echo "🔄 更新 Go 版本到 $(GO_VERSION)..."
	@echo "更新 go.work 文件..."
	@sed -i '' 's/^go .*/go $(GO_VERSION)/' go.work
	@echo "更新 .go-version 文件..."
	@echo "$(GO_VERSION)" > .go-version
	@echo "更新所有 go.mod 文件..."
	@for dir in $(MODULES); do \
		if [ "$$dir" = "." ]; then \
			echo "  - 更新根模块 go.mod"; \
			sed -i '' 's/^go .*/go $(GO_VERSION)/' go.mod; \
		else \
			echo "  - 更新模块 go.mod: $$dir"; \
			sed -i '' 's/^go .*/go $(GO_VERSION)/' $$dir/go.mod; \
		fi; \
	done
	@$(MAKE) update

benchmark: ## 运行性能测试
	@echo "⚡ 运行性能测试..."
	@for dir in $(MODULES); do \
		if [ "$$dir" = "." ]; then \
			echo "  - 测试根模块"; \
			go test -bench=. -benchmem ./...; \
		else \
			echo "  - 测试模块: $$dir"; \
			(cd $$dir && go test -bench=. -benchmem ./...); \
		fi; \
	done

coverage: ## 生成测试覆盖率报告
	@echo "📊 生成测试覆盖率报告..."
	@mkdir -p coverage
	@for dir in $(MODULES); do \
		if [ "$$dir" = "." ]; then \
			module_name="root"; \
		else \
			module_name="$$dir"; \
		fi; \
		echo "  - 生成模块覆盖率: $$module_name"; \
		if [ "$$dir" = "." ]; then \
			go test -coverprofile=coverage/$$module_name.out ./...; \
		else \
			(cd $$dir && go test -coverprofile=../coverage/$$module_name.out ./...); \
		fi; \
	done
	@echo "合并覆盖率报告..."
	@go tool covdata textfmt -i=coverage -o coverage/total.out 2>/dev/null || echo "使用简单合并方式"
	@go tool cover -html=coverage/root.out -o coverage/coverage.html
	@echo "覆盖率报告生成完成: coverage/coverage.html"

install-tools: ## 安装开发工具
	@echo "🔧 安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "开发工具安装完成"

docker-build: ## 构建 Docker 镜像（如果有 Dockerfile）
	@if [ -f Dockerfile ]; then \
		echo "🐳 构建 Docker 镜像..."; \
		docker build -t $(PROJECT_NAME):latest .; \
	else \
		echo "❌ 未找到 Dockerfile"; \
	fi

release-check: ## 发布前检查
	@echo "🚀 发布前检查..."
	@$(MAKE) fmt
	@$(MAKE) lint
	@$(MAKE) test
	@$(MAKE) verify
	@echo "✅ 发布前检查完成"

# 快捷命令别名
u: update        ## 别名：update
s: sync          ## 别名：sync
v: verify        ## 别名：verify
t: test          ## 别名：test
st: status       ## 别名：status 