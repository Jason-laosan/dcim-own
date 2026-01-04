# DCIM系统快速开始指南

## 目录
- [系统要求](#系统要求)
- [5分钟快速体验](#5分钟快速体验)
- [详细部署步骤](#详细部署步骤)
- [验证部署](#验证部署)
- [添加第一个采集任务](#添加第一个采集任务)
- [常见问题](#常见问题)

## 系统要求

### 最低配置
- CPU: 4核
- 内存: 8GB
- 磁盘: 50GB
- 操作系统: Windows 10/11, Linux, macOS

### 软件依赖
- Docker Desktop 20.10+
- Docker Compose 1.29+
- (可选) Go 1.21+（用于本地开发）
- (可选) Kubernetes 1.20+（用于生产部署）

## 5分钟快速体验

### Windows用户

1. **启动Docker Desktop**

2. **双击运行启动脚本**
   ```
   双击 start.bat 文件
   ```

3. **等待服务启动**（首次启动需要下载镜像，约3-5分钟）

4. **访问服务**
   - EMQX Dashboard: http://localhost:18083 (用户名: admin, 密码: public)
   - InfluxDB UI: http://localhost:8086
   - MinIO Console: http://localhost:9001 (用户名: minioadmin, 密码: minioadmin)

### Linux/macOS用户

```bash
# 1. 启动所有服务
docker-compose up -d

# 2. 查看服务状态
docker-compose ps

# 3. 查看日志
docker-compose logs -f
```

## 详细部署步骤

### 步骤1: 克隆/下载项目

```bash
# 如果使用Git
git clone <repository-url>
cd dcim-system

# 或直接解压下载的ZIP文件
```

### 步骤2: 初始化配置文件

```bash
# Windows
copy collector-agent\config.yaml.example collector-agent\config.yaml
copy services\collector-mgmt\config.yaml.example services\collector-mgmt\config.yaml
copy services\data-processor\config.yaml.example services\data-processor\config.yaml

# Linux/macOS
cp collector-agent/config.yaml.example collector-agent/config.yaml
cp services/collector-mgmt/config.yaml.example services/collector-mgmt/config.yaml
cp services/data-processor/config.yaml.example services/data-processor/config.yaml
```

### 步骤3: 修改配置（可选）

如果需要修改默认配置，编辑以下文件：

**docker-compose.yml** - 修改端口映射、环境变量等

**collector-agent/config.yaml** - 修改Agent配置
- `agent.id`: Agent唯一标识
- `mqtt.broker`: MQTT服务器地址
- `mqtt.topic`: 数据上报主题

**services/collector-mgmt/config.yaml** - 修改采集管理服务配置

**services/data-processor/config.yaml** - 修改数据处理服务配置

### 步骤4: 启动基础设施

```bash
# 仅启动基础设施（不启动应用服务）
docker-compose up -d emqx redis influxdb postgres minio

# 等待服务就绪（约30秒）
sleep 30
```

### 步骤5: 初始化InfluxDB

访问 http://localhost:8086 完成InfluxDB初始化：
1. 设置用户名和密码
2. 组织名称: dcim
3. Bucket名称: device_metrics
4. 复制生成的Token，更新到 `services/data-processor/config.yaml` 中

### 步骤6: 启动应用服务

```bash
# 启动所有应用服务
docker-compose up -d collector-mgmt data-processor

# 或者启动所有服务
docker-compose up -d
```

## 验证部署

### 1. 检查服务状态

```bash
docker-compose ps
```

所有服务应该显示为 "Up" 状态。

### 2. 检查服务日志

```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs collector-mgmt
docker-compose logs data-processor
```

### 3. 访问管理界面

#### EMQX Dashboard
- URL: http://localhost:18083
- 用户名: admin
- 密码: public
- 功能: 查看MQTT连接、消息统计

#### InfluxDB UI
- URL: http://localhost:8086
- 功能: 查看时序数据、执行查询

#### MinIO Console
- URL: http://localhost:9001
- 用户名: minioadmin
- 密码: minioadmin
- 功能: 对象存储管理

### 4. 测试API

```bash
# 健康检查
curl http://localhost:8080/health

# 预期返回: {"status":"ok"}
```

## 添加第一个采集任务

### 1. 准备设备信息

假设你有一台支持SNMP的网络设备：
- IP: 192.168.1.100
- SNMP版本: v2c
- Community: public

### 2. 创建采集任务

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-001",
    "task_id": "task-switch-001",
    "device_id": "switch-001",
    "device_ip": "192.168.1.100",
    "device_type": "switch",
    "protocol": "snmp",
    "interval": 30,
    "metrics": [
      "1.3.6.1.2.1.1.3.0",
      "1.3.6.1.2.1.2.2.1.10",
      "1.3.6.1.2.1.2.2.1.16"
    ],
    "config": {
      "version": "v2c",
      "community": "public",
      "port": 161,
      "timeout": 5
    }
  }'
```

### 3. 查询任务状态

```bash
curl "http://localhost:8080/api/v1/tasks?agent_id=agent-001"
```

### 4. 在InfluxDB中查询数据

访问 http://localhost:8086，进入Data Explorer，执行查询：

```flux
from(bucket: "device_metrics")
  |> range(start: -1h)
  |> filter(fn: (r) => r["device_id"] == "switch-001")
```

## 常见问题

### Q1: Docker服务启动失败

**解决方案：**
```bash
# 检查端口是否被占用
netstat -ano | findstr "1883"  # Windows
lsof -i :1883                  # Linux/macOS

# 检查Docker磁盘空间
docker system df

# 清理未使用的镜像
docker system prune -a
```

### Q2: MQTT连接失败

**解决方案：**
1. 检查EMQX是否正常运行：`docker-compose logs emqx`
2. 验证配置文件中的MQTT地址是否正确
3. 检查防火墙是否阻止了1883端口

### Q3: InfluxDB写入失败

**解决方案：**
1. 确认Token是否配置正确
2. 检查Bucket名称是否匹配
3. 查看data-processor日志：`docker-compose logs data-processor`

### Q4: Agent无法连接gRPC服务

**解决方案：**
1. 确认collector-mgmt服务是否启动
2. 检查gRPC端口50051是否可访问
3. 验证TLS配置（如果启用）

### Q5: 如何重启某个服务

```bash
# 重启单个服务
docker-compose restart collector-mgmt

# 重新构建并启动
docker-compose up -d --build collector-mgmt
```

### Q6: 如何查看实时日志

```bash
# 查看所有服务实时日志
docker-compose logs -f

# 查看特定服务实时日志
docker-compose logs -f data-processor

# 查看最近100行日志
docker-compose logs --tail=100 collector-mgmt
```

### Q7: 如何完全清理并重新开始

```bash
# 停止并删除所有容器和数据卷
docker-compose down -v

# 重新启动
docker-compose up -d
```

## 下一步

- 📖 阅读[完整文档](README.md)了解系统架构
- 🔧 学习如何[添加自定义协议插件](docs/protocol-plugin.md)
- 📊 配置[监控大屏](docs/dashboard.md)
- ⚙️ 进行[生产环境部署](docs/production-deploy.md)

## 获取帮助

- 提交Issue: [GitHub Issues]
- 查看文档: [完整文档](README.md)
- 联系团队: [联系方式]

---

祝使用愉快！
