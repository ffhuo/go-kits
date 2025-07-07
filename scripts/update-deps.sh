#!/bin/bash

# Go-Kits 依赖自动更新脚本
# 用于更新工作区中所有模块的依赖关系

set -e

# 获取脚本所在目录的父目录（项目根目录）
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo "🚀 开始更新 go-kits 项目依赖..."

# 定义所有模块目录
MODULES=(
    "."
    "core"
    "logger"
    "config"
    "redis"
    "logger_v2"
    "excel"
    "message"
    "storage"
    "influxdb"
    "etcd"
    "prometheus"
    "mqtt"
    "gout"
    "sqldb"
)

# 函数：更新单个模块
update_module() {
    local module_dir="$1"
    echo "📦 更新模块: $module_dir"
    
    if [ "$module_dir" = "." ]; then
        echo "  - 更新根模块"
    else
        cd "$module_dir"
    fi
    
    # 运行 go mod tidy
    if go mod tidy; then
        echo "  ✅ $module_dir 更新成功"
    else
        echo "  ❌ $module_dir 更新失败"
        return 1
    fi
    
    # 返回项目根目录
    cd "$PROJECT_ROOT"
}

# 函数：同步工作区
sync_workspace() {
    echo "🔄 同步工作区..."
    if go work sync; then
        echo "  ✅ 工作区同步成功"
    else
        echo "  ❌ 工作区同步失败"
        return 1
    fi
}

# 函数：验证构建
verify_build() {
    echo "🔍 验证构建..."
    if go build ./...; then
        echo "  ✅ 所有模块构建成功"
    else
        echo "  ❌ 构建失败"
        return 1
    fi
}

# 函数：运行测试
run_tests() {
    echo "🧪 运行测试..."
    local failed_modules=()
    
    for module in "${MODULES[@]}"; do
        echo "  - 测试模块: $module"
        if [ "$module" = "." ]; then
            if ! go test -timeout 30s ./...; then
                failed_modules+=("root")
            fi
        else
            if ! (cd "$module" && go test -timeout 30s ./...); then
                failed_modules+=("$module")
            fi
        fi
    done
    
    if [ ${#failed_modules[@]} -eq 0 ]; then
        echo "  ✅ 所有测试通过"
    else
        echo "  ⚠️ 以下模块测试失败: ${failed_modules[*]}"
    fi
}

# 函数：显示依赖状态
show_status() {
    echo "📊 依赖状态概览:"
    echo "===================="
    
    for module in "${MODULES[@]}"; do
        if [ "$module" = "." ]; then
            echo "📁 根模块"
            go_version=$(grep "^go " go.mod | awk '{print $2}')
            echo "  Go 版本: $go_version"
        else
            echo "📁 $module"
            if [ -f "$module/go.mod" ]; then
                go_version=$(grep "^go " "$module/go.mod" | awk '{print $2}')
                echo "  Go 版本: $go_version"
                
                # 显示内部依赖
                internal_deps=$(grep "github.com/ffhuo/go-kits" "$module/go.mod" | grep -v "module" | wc -l)
                if [ "$internal_deps" -gt 0 ]; then
                    echo "  内部依赖: $internal_deps 个"
                fi
            else
                echo "  ❌ 缺少 go.mod 文件"
            fi
        fi
        echo ""
    done
}

# 主执行流程
main() {
    case "${1:-update}" in
        "update")
            echo "🔄 执行完整更新流程..."
            
            # 1. 同步工作区
            sync_workspace
            
            # 2. 更新所有模块
            for module in "${MODULES[@]}"; do
                update_module "$module"
            done
            
            # 3. 再次同步工作区
            sync_workspace
            
            # 4. 验证构建
            verify_build
            
            echo "🎉 依赖更新完成！"
            ;;
            
        "sync")
            echo "🔄 仅同步工作区..."
            sync_workspace
            ;;
            
        "verify")
            echo "🔍 验证项目状态..."
            verify_build
            ;;
            
        "test")
            echo "🧪 运行所有测试..."
            run_tests
            ;;
            
        "status")
            echo "📊 显示依赖状态..."
            show_status
            ;;
            
        "help"|"-h"|"--help")
            echo "Go-Kits 依赖管理脚本"
            echo ""
            echo "用法: $0 [命令]"
            echo ""
            echo "命令:"
            echo "  update  (默认) 完整更新所有模块依赖"
            echo "  sync           仅同步工作区"
            echo "  verify         验证所有模块可以构建"
            echo "  test           运行所有模块的测试"
            echo "  status         显示当前依赖状态"
            echo "  help           显示此帮助信息"
            echo ""
            echo "示例:"
            echo "  $0              # 完整更新"
            echo "  $0 sync         # 仅同步"
            echo "  $0 status       # 查看状态"
            ;;
            
        *)
            echo "❌ 未知命令: $1"
            echo "使用 '$0 help' 查看可用命令"
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@" 