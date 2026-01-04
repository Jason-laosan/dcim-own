.PHONY: help build build-agent build-services docker-build docker-up docker-down clean test

# 默认目标
help:
	@echo "DCIM系统构建工具"
	@echo ""
	@echo "使用方法："
	@echo "  make build              - 构建所有组件"
	@echo "  make build-agent        - 构建采集Agent"
	@echo "  make build-services     - 构建所有微服务"
	@echo "  make docker-build       - 构建Docker镜像"
	@echo "  make docker-up          - 启动Docker Compose服务"
	@echo "  make docker-down        - 停止Docker Compose服务"
	@echo "  make test               - 运行测试"
	@echo "  make clean              - 清理构建文件"

# 构建所有组件
build: build-agent build-services

# 构建采集Agent
build-agent:
	@echo "构建采集Agent..."
	cd collector-agent && go build -o bin/collector-agent cmd/main.go
	@echo "采集Agent构建完成: collector-agent/bin/collector-agent"

# 构建所有微服务
build-services:
	@echo "构建采集管理服务..."
	cd services/collector-mgmt && go build -o bin/collector-mgmt cmd/main.go
	@echo "构建数据处理服务..."
	cd services/data-processor && go build -o bin/data-processor cmd/main.go
	@echo "所有微服务构建完成"

# 构建Docker镜像
docker-build:
	@echo "构建Docker镜像..."
	docker build -f deploy/docker/Dockerfile.collector-agent -t dcim/collector-agent:latest .
	docker build -f deploy/docker/Dockerfile.collector-mgmt -t dcim/collector-mgmt:latest ./services/collector-mgmt
	docker build -f deploy/docker/Dockerfile.data-processor -t dcim/data-processor:latest ./services/data-processor
	@echo "Docker镜像构建完成"

# 启动Docker Compose服务
docker-up:
	@echo "启动Docker Compose服务..."
	docker-compose up -d
	@echo "服务启动完成"
	@echo "EMQX Dashboard: http://localhost:18083"
	@echo "InfluxDB UI: http://localhost:8086"
	@echo "MinIO Console: http://localhost:9001"
	@echo "采集管理服务: http://localhost:8080"

# 停止Docker Compose服务
docker-down:
	@echo "停止Docker Compose服务..."
	docker-compose down
	@echo "服务已停止"

# 运行测试
test:
	@echo "运行测试..."
	cd collector-agent && go test ./...
	cd services/collector-mgmt && go test ./...
	cd services/data-processor && go test ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf collector-agent/bin
	rm -rf services/collector-mgmt/bin
	rm -rf services/data-processor/bin
	@echo "清理完成"

# 初始化配置文件
init-config:
	@echo "初始化配置文件..."
	cd collector-agent && cp config.yaml.example config.yaml
	cd services/collector-mgmt && cp config.yaml.example config.yaml
	cd services/data-processor && cp config.yaml.example config.yaml
	@echo "配置文件初始化完成，请根据实际情况修改配置"
