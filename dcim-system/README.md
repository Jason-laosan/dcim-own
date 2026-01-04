# DCIM数据中心基础设施管理系统

## 项目简介

DCIM（Data Center Infrastructure Management）系统是一个完整的数据中心基础设施管理平台，专注于机柜、UPS、精密空调、PDU、温湿度传感器、服务器/网络设备等基础设施的全生命周期管理。

核心能力：
- **实时监控**：秒级/分钟级设备数据采集
- **故障告警**：实时告警检测和分级推送
- **资产盘点**：设备资产全生命周期管理
- **能耗分析**：按机房/机柜/设备维度统计分析
- **容量规划**：U位/功率/空间容量预警

## 技术架构

本系统采用分层+微服务架构，共分为6层：

```
┌─────────────────────────────────────────────────────────────┐
│                     展示层 (Vue3)                             │
│   Web管理端 | 监控大屏 | 移动端 | 报表中心                     │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│              应用层 (API网关 + 权限中心)                        │
│   Nginx + Kong | RBAC权限 | 任务调度                          │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                  计算层 (微服务)                               │
│ 采集管理 | 数据处理 | 资产管理 | 告警 | 能耗分析 | 容量规划      │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    存储层                                      │
│   InfluxDB | PostgreSQL | Redis | MinIO                      │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                    传输层                                      │
│         MQTT (EMQ X) | gRPC | HTTP/HTTPS                     │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│              采集层 (Golang Agent)                             │
│     多协议采集 | 边缘计算 | 任务调度 | 本地缓存                 │
└─────────────────────────────────────────────────────────────┘
```

## 核心技术栈

| 层级 | 技术选型 |
|------|---------|
| 采集层 | Golang + SNMP/IPMI/Modbus/HTTP |
| 传输层 | EMQ X (MQTT) + gRPC |
| 存储层 | InfluxDB + PostgreSQL + Redis + MinIO |
| 计算层 | Golang (Gin) + Java (Spring Boot) |
| 展示层 | Vue3 + Element Plus + ECharts |
| 容器化 | Docker + Kubernetes |

## 项目结构

```
dcim-system/
├── collector-agent/              # 采集Agent（Golang）
│   ├── cmd/                      # 主程序入口
│   ├── internal/                 # 内部模块
│   │   ├── agent/                # Agent核心
│   │   ├── cache/                # 本地缓存
│   │   ├── collector/            # 数据采集器
│   │   ├── protocol/             # 协议插件
│   │   ├── scheduler/            # 任务调度
│   │   └── heartbeat/            # 心跳模块
│   ├── pkg/                      # 公共包
│   │   ├── config/               # 配置管理
│   │   ├── logger/               # 日志管理
│   │   └── utils/                # 工具函数
│   └── config.yaml.example       # 配置示例
│
├── services/                     # 微服务
│   ├── collector-mgmt/           # 采集管理服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── handler/          # HTTP处理器
│   │   │   ├── service/          # 业务逻辑
│   │   │   └── repository/       # 数据访问
│   │   └── pkg/
│   │
│   ├── data-processor/           # 数据处理服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── processor/        # 数据处理
│   │   │   └── storage/          # 存储逻辑
│   │   └── pkg/
│   │
│   ├── asset-mgmt/               # 资产管理服务（待开发）
│   ├── alert/                    # 告警服务（待开发）
│   ├── energy/                   # 能耗分析服务（待开发）
│   └── capacity/                 # 容量规划服务（待开发）
│
├── proto/                        # gRPC协议定义
│   └── collector.proto
│
├── deploy/                       # 部署文件
│   ├── docker/                   # Docker文件
│   │   ├── Dockerfile.collector-agent
│   │   ├── Dockerfile.collector-mgmt
│   │   └── Dockerfile.data-processor
│   └── k8s/                      # Kubernetes配置
│       ├── collector-mgmt.yaml
│       └── data-processor.yaml
│
├── configs/                      # 配置文件
├── scripts/                      # 脚本工具
└── docker-compose.yml            # Docker Compose配置
```

## 快速开始

### 前置要求

- Go 1.21+
- Docker & Docker Compose
- Kubernetes (可选，用于生产部署)

### 本地开发环境

1. **克隆项目**
```bash
git clone <repository-url>
cd dcim-system
```

2. **启动基础设施（使用Docker Compose）**
```bash
docker-compose up -d emqx redis influxdb postgres minio
```

3. **配置服务**

采集Agent配置：
```bash
cd collector-agent
cp config.yaml.example config.yaml
# 编辑config.yaml，修改MQTT、gRPC等配置
```

采集管理服务配置：
```bash
cd services/collector-mgmt
cp config.yaml.example config.yaml
# 编辑config.yaml，修改Redis等配置
```

数据处理服务配置：
```bash
cd services/data-processor
cp config.yaml.example config.yaml
# 编辑config.yaml，修改MQTT、InfluxDB等配置
```

4. **运行服务**

运行采集Agent：
```bash
cd collector-agent
go run cmd/main.go -config config.yaml
```

运行采集管理服务：
```bash
cd services/collector-mgmt
go run cmd/main.go -config config.yaml
```

运行数据处理服务：
```bash
cd services/data-processor
go run cmd/main.go -config config.yaml
```

### Docker Compose部署（推荐）

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

访问服务：
- EMQX Dashboard: http://localhost:18083 (用户名: admin, 密码: public)
- InfluxDB UI: http://localhost:8086
- MinIO Console: http://localhost:9001
- 采集管理服务API: http://localhost:8080

### Kubernetes部署

1. **创建命名空间**
```bash
kubectl create namespace dcim
```

2. **部署服务**
```bash
# 部署采集管理服务
kubectl apply -f deploy/k8s/collector-mgmt.yaml

# 部署数据处理服务
kubectl apply -f deploy/k8s/data-processor.yaml
```

3. **查看状态**
```bash
kubectl get pods -n dcim
kubectl get svc -n dcim
```

## 核心功能

### 1. 设备采集

支持的协议：
- **SNMP (v1/v2c/v3)**: 网络设备、PDU、UPS、空调
- **IPMI**: 物理服务器
- **Modbus (TCP/RTU)**: 温湿度传感器、智能电表
- **HTTP/RESTful**: 智能PDU、云化设备
- **SSH/Telnet**: 服务器/网络设备命令行采集

### 2. 数据流转

1. Agent按配置的频率采集设备数据
2. 数据经过预处理（清洗、格式转换）
3. 通过MQTT上报至消息队列（QoS=1）
4. 数据处理服务消费消息
5. 写入InfluxDB（时序数据）和Redis（热点数据）
6. 前端大屏实时展示

### 3. API接口

#### 添加采集任务
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-001",
    "task_id": "task-001",
    "device_id": "device-001",
    "device_ip": "192.168.1.100",
    "device_type": "switch",
    "protocol": "snmp",
    "interval": 30,
    "metrics": ["1.3.6.1.2.1.1.3.0"],
    "config": {
      "version": "v2c",
      "community": "public"
    }
  }'
```

#### 查询Agent状态
```bash
curl http://localhost:8080/api/v1/agent/status?agent_id=agent-001
```

#### 查询任务列表
```bash
curl http://localhost:8080/api/v1/tasks?agent_id=agent-001
```

## 性能指标

- 单Agent并发采集能力：≥1000台设备
- 采集频率：最低1秒
- 数据上报延迟：＜100ms
- Agent体积：＜10MB（单二进制文件）
- 内存占用：＜100MB（正常负载）

## 开发指南

### 添加新协议插件

1. 在`collector-agent/internal/protocol/`目录下创建新协议实现
2. 实现`Protocol`接口
3. 在Agent中注册协议

示例：
```go
// 实现HTTP协议插件
type HTTPProtocol struct {
    client *http.Client
}

func (h *HTTPProtocol) Collect(ctx context.Context, task *CollectTask) (*DeviceData, error) {
    // 实现采集逻辑
}
```

### 构建二进制文件

```bash
# 构建采集Agent
cd collector-agent
go build -o bin/collector-agent cmd/main.go

# 构建采集管理服务
cd services/collector-mgmt
go build -o bin/collector-mgmt cmd/main.go

# 构建数据处理服务
cd services/data-processor
go build -o bin/data-processor cmd/main.go
```

## 监控与告警

系统自带监控指标：
- Agent状态监控（心跳检测）
- 采集任务执行情况
- MQTT消息队列状态
- 数据库连接状态

## 安全设计

- Agent与服务端通信加密（TLS 1.3）
- MQTT认证（用户名+密码）
- 数据传输加密
- 接口鉴权（JWT Token）
- 敏感数据加密存储

## 贡献指南

欢迎提交Issue和Pull Request！

## 许可证

[MIT License](LICENSE)

## 联系方式

如有问题，请提交Issue或联系维护团队。

---

**注意**：本项目基于架构设计文档生成，部分功能仍在开发中。
