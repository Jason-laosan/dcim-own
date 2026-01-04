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

本系统采用分层+微服务架构，数据采集层基于Golang开发，充分利用其高并发、轻量级、跨平台特性。

### 核心技术栈

| 层级 | 技术选型 |
|------|---------|
| 采集层 | Golang + SNMP/IPMI/Modbus/HTTP |
| 传输层 | EMQ X (MQTT) + gRPC |
| 存储层 | InfluxDB + PostgreSQL + Redis + MinIO |
| 计算层 | Golang (Gin) + Java (Spring Boot) |
| 展示层 | Vue3 + Element Plus + ECharts |
| 容器化 | Docker + Kubernetes |

## 架构分层

系统分为6层，实现"采集-传输-存储-计算-应用-展示"全链路设计：

1. **采集层（Golang）**：多协议设备数据采集、边缘计算、任务调度
2. **传输层**：数据可靠传输、协议转换、流量控制
3. **存储层**：时序数据、结构化数据、缓存、文件存储
4. **计算层**：数据处理、告警引擎、资产管理、能耗分析
5. **应用层**：API网关、权限控制、任务调度
6. **展示层**：Web管理端、监控大屏、移动端、报表中心

## 支持的采集协议

- **SNMP (v1/v2c/v3)**：网络设备、PDU、UPS、空调
- **IPMI**：物理服务器
- **Modbus (TCP/RTU)**：温湿度传感器、智能电表
- **HTTP/RESTful**：智能PDU、云化设备
- **SSH/Telnet**：服务器/网络设备命令行采集
- **BACnet/IP**：楼宇自控、空调系统

## 项目结构

```
dcim-own/
├── dcim-system/              # 主系统代码
│   ├── collector-agent/      # 采集Agent（Golang）
│   ├── services/             # 微服务
│   ├── proto/                # gRPC协议定义
│   ├── deploy/               # 部署文件
│   └── configs/              # 配置文件
└── DCIM系统架构设计.md       # 详细架构设计文档
```

## 快速开始

详细的部署和开发指南请参考：
- [系统架构设计文档](./DCIM（数据中心基础设施管理）系统架构设计（数据采集层基于Golang）.md)
- [项目开发指南](./dcim-system/README.md)
- [快速入门指南](./dcim-system/QUICKSTART.md)

### 前置要求

- Go 1.21+
- Docker & Docker Compose
- Kubernetes（可选，用于生产部署）

### Docker Compose快速部署

```bash
cd dcim-system
docker-compose up -d
```

访问服务：
- EMQX Dashboard: http://localhost:18083
- InfluxDB UI: http://localhost:8086
- MinIO Console: http://localhost:9001

## 核心特性

### 高性能采集
- 单Agent并发采集能力：≥1000台设备
- 采集频率：最低1秒
- 数据上报延迟：＜100ms
- Agent体积：＜10MB（单二进制文件）

### 高可用设计
- Agent本地缓存 + 断线重连
- MQTT Broker集群
- 微服务集群部署
- 数据库主从复制 + 定时备份

### 安全设计
- Agent与服务端通信加密（TLS 1.3）
- MQTT认证（用户名+密码）
- 接口鉴权（JWT Token）
- 敏感数据加密存储

## 开发指南

### 构建项目

```bash
# 构建采集Agent
cd dcim-system/collector-agent
go build -o bin/collector-agent cmd/main.go

# 构建采集管理服务
cd dcim-system/services/collector-mgmt
go build -o bin/collector-mgmt cmd/main.go
```

### 添加新协议插件

在`collector-agent/internal/protocol/`目录下创建新协议实现，实现`Protocol`接口即可。

## 贡献指南

欢迎提交Issue和Pull Request！

1. Fork本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启Pull Request

## 许可证

本项目采用MIT许可证 - 详见 [LICENSE](LICENSE) 文件

## 联系方式

如有问题，请提交Issue或联系维护团队。

## 致谢

本项目基于数据中心基础设施管理最佳实践开发，感谢开源社区的贡献。

---

**注意**：本项目部分功能仍在开发中，欢迎参与贡献。
