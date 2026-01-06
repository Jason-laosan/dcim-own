#!/bin/bash

echo "=== TaskHub - Go 学习项目 ==="
echo "正在启动..."
echo ""

cd "$(dirname "$0")"

# 检查 Go 是否安装
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go。请先安装 Go 1.21 或更高版本。"
    exit 1
fi

# 运行项目
go run cmd/server/main.go
