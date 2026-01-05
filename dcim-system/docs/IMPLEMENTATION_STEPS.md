# DCIM双模式采集系统实施步骤

## 目录
- [环境准备](#环境准备)
- [基础设施部署](#基础设施部署)
- [Agent配置与启动](#agent配置与启动)
- [主动拉取模式配置](#主动拉取模式配置)
- [被动接收模式配置](#被动接收模式配置)
- [验证与测试](#验证与测试)
- [生产部署](#生产部署)
- [故障排查](#故障排查)

---

## 环境准备

### 1. 安装Go环境

```bash
# 检查Go版本（需要1.21+）
go version

# 如果未安装，下载安装Go 1.21+
# Windows: https://go.dev/dl/
# Linux: 
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### 2. 克隆项目

```bash
cd /path/to/your/workspace
git clone <repository-url>
cd dcim-system
```

### 3. 安装依赖

```bash
# 进入collector-agent目录
cd collector-agent

# 下载并整理依赖
go mod tidy

# 验证依赖
go mod verify
```

---

## 基础设施部署

### 方式1：使用Docker Compose（推荐用于开发/测试）

```bash
# 返回项目根目录
cd ..

# 启动基础设施
docker-compose up -d emqx redis influxdb postgres minio

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f emqx
```

**服务访问地址**：
- EMQX Dashboard: http://localhost:18083 (admin/public)
- InfluxDB UI: http://localhost:8086
- MinIO Console: http://localhost:9001 (minioadmin/minioadmin)
- Redis: localhost:6379
- PostgreSQL: localhost:5432

### 方式2：手动安装

#### 安装EMQX（MQTT Broker）

```bash
# Ubuntu/Debian
wget https://www.emqx.io/downloads/broker/v5.3.0/emqx-5.3.0-ubuntu20.04-amd64.deb
sudo dpkg -i emqx-5.3.0-ubuntu20.04-amd64.deb
sudo systemctl start emqx
sudo systemctl enable emqx

# 验证
curl http://localhost:18083
```

#### 安装Redis

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install redis-server

# 启动
sudo systemctl start redis
sudo systemctl enable redis

# 验证
redis-cli ping
```

#### 安装InfluxDB

```bash
# Ubuntu/Debian
wget https://dl.influxdata.com/influxdb/releases/influxdb2-2.7.0-amd64.deb
sudo dpkg -i influxdb2-2.7.0-amd64.deb
sudo systemctl start influxdb
sudo systemctl enable influxdb

# 初始化
influx setup
```

---

## Agent配置与启动

### 1. 选择配置模式

根据您的需求选择配置模板：

```bash
cd collector-agent

# 方案A: 双模式（推荐）- 同时支持主动拉取和被动接收
cp config-dual-mode.yaml.example config.yaml

# 方案B: 仅主动拉取 - 适用于传统设备
cp config-pull-only.yaml.example config.yaml

# 方案C: 仅被动接收 - 适用于智能设备
cp config-push-only.yaml.example config.yaml
```

### 2. 编辑配置文件

```bash
# 使用您喜欢的编辑器打开配置文件
vim config.yaml
# 或
code config.yaml
```

**必须修改的配置项**：

```yaml
agent:
  id: "agent-001"                    # 修改为唯一的Agent ID
  name: "您的数据中心名称"           # 修改为实际名称
  data_center: "DC-01"               # 数据中心编号
  room: "Room-A"                     # 机房编号
  enable_pull_mode: true             # 是否启用主动拉取
  enable_push_mode: true             # 是否启用被动接收

mqtt:
  broker: "tcp://localhost:1883"     # 修改为实际MQTT地址
  username: "dcim_agent"             # MQTT用户名
  password: "your_password"          # MQTT密码
  topic: "dcim/collector/data"       # 数据上报Topic
  client_id: "agent-001"             # 客户端ID

grpc:
  server_addr: "localhost:50051"     # 采集管理服务地址

receiver:
  enabled: true                      # 启用被动接收
  mqtt_receiver:
    enabled: true
    broker: "tcp://localhost:1883"
    subscribe_topics:
      - "device/+/data"              # 修改为实际订阅Topic
      - "sensor/+/metrics"
```

### 3. 编译Agent

```bash
# 编译
go build -o bin/collector-agent cmd/main.go

# 或使用Makefile（如果有）
make build-agent

# Windows编译
go build -o bin/collector-agent.exe cmd/main.go
```

### 4. 启动Agent

```bash
# 前台运行（用于测试）
./bin/collector-agent -config config.yaml

# 后台运行（Linux）
nohup ./bin/collector-agent -config config.yaml > logs/agent.log 2>&1 &

# 使用systemd管理（推荐生产环境）
sudo cp deploy/systemd/collector-agent.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start collector-agent
sudo systemctl enable collector-agent
```

### 5. 验证启动

查看日志确认Agent正常启动：

```bash
# 查看日志
tail -f logs/agent.log

# 应该看到以下日志
[INFO] config loaded agent_id=agent-001
[INFO] MQTT connected
[INFO] pull mode started          # 如果启用了Pull模式
[INFO] push mode started          # 如果启用了Push模式
[INFO] MQTT receiver connected    # 如果启用了MQTT接收器
[INFO] agent started successfully
```

---

## 主动拉取模式配置

### 1. 启动采集管理服务

```bash
# 进入采集管理服务目录
cd ../services/collector-mgmt

# 配置服务
cp config.yaml.example config.yaml
vim config.yaml

# 编译并启动
go build -o bin/collector-mgmt cmd/main.go
./bin/collector-mgmt -config config.yaml
```

### 2. 添加SNMP采集任务

```bash
# 添加交换机SNMP采集任务
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-001",
    "task_id": "task-switch-001",
    "device_id": "switch-001",
    "device_ip": "192.168.1.100",
    "device_type": "switch",
    "protocol": "snmp",
    "mode": "pull",
    "interval": 30,
    "metrics": [
      "1.3.6.1.2.1.1.3.0",
      "1.3.6.1.2.1.2.2.1.10.1",
      "1.3.6.1.2.1.2.2.1.16.1"
    ],
    "config": {
      "version": "v2c",
      "community": "public",
      "port": 161,
      "timeout": 5
    }
  }'
```

### 3. 添加HTTP采集任务

```bash
# 添加智能PDU HTTP采集任务
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-001",
    "task_id": "task-pdu-001",
    "device_id": "pdu-001",
    "device_ip": "192.168.1.200",
    "device_type": "pdu",
    "protocol": "http",
    "mode": "pull",
    "interval": 60,
    "metrics": ["/api/v1/metrics"],
    "config": {
      "url": "http://192.168.1.200/api/v1/metrics",
      "method": "GET",
      "timeout": 10
    }
  }'
```

### 4. 查询任务列表

```bash
# 查询Agent的所有任务
curl http://localhost:8080/api/v1/tasks?agent_id=agent-001

# 查询特定任务
curl http://localhost:8080/api/v1/tasks/task-switch-001
```

### 5. 删除任务

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/task-switch-001
```

---

## 被动接收模式配置

### 1. 配置MQTT接收器

确保 `config.yaml` 中MQTT接收器已启用：

```yaml
receiver:
  enabled: true
  mqtt_receiver:
    enabled: true
    broker: "tcp://localhost:1883"
    username: "dcim_receiver"
    password: "your_password"
    subscribe_topics:
      - "device/+/data"        # 订阅所有设备数据
      - "sensor/+/metrics"     # 订阅传感器指标
      - "smart_pdu/+/status"   # 订阅智能PDU状态
    qos: 1
    client_id: "agent-001-receiver"
```

### 2. 设备推送数据示例

#### Python示例（MQTT推送）

```python
# install: pip install paho-mqtt
import paho.mqtt.client as mqtt
import json
import time

# 连接MQTT Broker
client = mqtt.Client()
client.username_pw_set("device_user", "device_password")
client.connect("localhost", 1883, 60)

# 推送设备数据
def push_device_data():
    data = {
        "device_id": "sensor-001",
        "device_ip": "192.168.1.50",
        "device_type": "temperature_sensor",
        "timestamp": time.time(),
        "metrics": {
            "temperature": 25.5,
            "humidity": 60.2
        },
        "status": "success",
        "error": ""
    }
    
    # 发布到Topic
    client.publish("sensor/sensor-001/metrics", json.dumps(data))
    print(f"Data pushed: {data}")

# 定时推送
while True:
    push_device_data()
    time.sleep(10)  # 每10秒推送一次
```

#### Node.js示例（MQTT推送）

```javascript
// install: npm install mqtt
const mqtt = require('mqtt');

// 连接MQTT Broker
const client = mqtt.connect('mqtt://localhost:1883', {
  username: 'device_user',
  password: 'device_password'
});

client.on('connect', () => {
  console.log('Connected to MQTT Broker');
  
  // 定时推送数据
  setInterval(() => {
    const data = {
      device_id: 'smart-pdu-001',
      device_ip: '192.168.1.200',
      device_type: 'pdu',
      timestamp: Date.now() / 1000,
      metrics: {
        voltage: 220.5,
        current: 15.2,
        power: 3350.0
      },
      status: 'success',
      error: ''
    };
    
    client.publish('smart_pdu/smart-pdu-001/status', JSON.stringify(data));
    console.log('Data pushed:', data);
  }, 5000); // 每5秒推送一次
});
```

### 3. 配置Modbus接收器

```yaml
receiver:
  modbus_receiver:
    enabled: true
    mode: "tcp"                # TCP模式
    listen_addr: "0.0.0.0:502" # 监听地址
    slave_id: 1                # 从站ID
```

### 4. 验证数据接收

```bash
# 订阅MQTT Topic查看Agent上报的数据
mosquitto_sub -h localhost -t "dcim/collector/data" -v

# 查看Agent日志
tail -f logs/agent.log | grep "received data from push mode"
```

---

## 验证与测试

### 1. 检查Agent状态

```bash
# 查询Agent状态
curl http://localhost:8080/api/v1/agent/status?agent_id=agent-001

# 订阅心跳Topic
mosquitto_sub -h localhost -t "dcim/collector/data/heartbeat"
```

### 2. 验证Pull模式

```bash
# 1. 添加测试任务
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-001",
    "task_id": "test-task-001",
    "device_id": "test-device-001",
    "device_ip": "192.168.1.1",
    "protocol": "snmp",
    "mode": "pull",
    "interval": 10,
    "metrics": ["1.3.6.1.2.1.1.3.0"],
    "config": {"version": "v2c", "community": "public"}
  }'

# 2. 查看采集日志
tail -f logs/agent.log | grep "collect success"

# 3. 订阅数据Topic查看上报数据
mosquitto_sub -h localhost -t "dcim/collector/data"
```

### 3. 验证Push模式

```bash
# 1. 使用mosquitto_pub模拟设备推送
mosquitto_pub -h localhost -t "device/test-001/data" -m '{
  "device_id": "test-001",
  "device_ip": "192.168.1.100",
  "device_type": "sensor",
  "timestamp": 1704438000,
  "metrics": {"temperature": 25.5},
  "status": "success"
}'

# 2. 查看接收日志
tail -f logs/agent.log | grep "received data from push mode"

# 3. 验证数据上报
mosquitto_sub -h localhost -t "dcim/collector/data"
```

### 4. 性能测试

```bash
# 批量添加任务测试并发采集
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{
      \"agent_id\": \"agent-001\",
      \"task_id\": \"task-$i\",
      \"device_id\": \"device-$i\",
      \"device_ip\": \"192.168.1.$i\",
      \"protocol\": \"snmp\",
      \"mode\": \"pull\",
      \"interval\": 30,
      \"metrics\": [\"1.3.6.1.2.1.1.3.0\"],
      \"config\": {\"version\": \"v2c\", \"community\": \"public\"}
    }"
done

# 监控Agent性能
top -p $(pgrep collector-agent)
```

---

## 生产部署

### 1. 使用Docker部署

```bash
# 构建Docker镜像
cd collector-agent
docker build -t dcim-collector-agent:latest -f ../deploy/docker/Dockerfile.collector-agent .

# 运行容器
docker run -d \
  --name collector-agent \
  --network dcim-network \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/data:/app/data \
  dcim-collector-agent:latest
```

### 2. 使用Kubernetes部署

```bash
# 创建ConfigMap
kubectl create configmap agent-config --from-file=config.yaml

# 部署Agent
kubectl apply -f deploy/k8s/collector-agent.yaml

# 查看状态
kubectl get pods -l app=collector-agent
kubectl logs -f deployment/collector-agent
```

### 3. 配置监控

```bash
# 使用Prometheus监控
# 1. Agent暴露metrics端点
# 2. 配置Prometheus抓取

# prometheus.yml
scrape_configs:
  - job_name: 'dcim-agent'
    static_configs:
      - targets: ['localhost:9090']
```

### 4. 配置告警

```yaml
# alertmanager配置
groups:
  - name: dcim-agent
    rules:
      - alert: AgentDown
        expr: up{job="dcim-agent"} == 0
        for: 1m
        annotations:
          summary: "Agent {{ $labels.instance }} is down"
      
      - alert: HighCollectFailureRate
        expr: collect_failure_rate > 0.1
        for: 5m
        annotations:
          summary: "Collect failure rate > 10%"
```

---

## 故障排查

### 问题1: Agent启动失败

**检查步骤**：
```bash
# 1. 检查配置文件
./bin/collector-agent -config config.yaml --validate

# 2. 检查端口占用
netstat -tulpn | grep 1883

# 3. 查看详细日志
./bin/collector-agent -config config.yaml -log-level debug
```

### 问题2: Pull模式采集失败

**检查步骤**：
```bash
# 1. 测试设备连通性
ping 192.168.1.100

# 2. 测试SNMP端口
snmpwalk -v2c -c public 192.168.1.100 1.3.6.1.2.1.1.3.0

# 3. 查看采集日志
grep "collect failed" logs/agent.log
```

### 问题3: Push模式未接收数据

**检查步骤**：
```bash
# 1. 验证MQTT连接
mosquitto_sub -h localhost -t "#" -v

# 2. 检查Topic订阅
grep "subscribed to topic" logs/agent.log

# 3. 测试数据推送
mosquitto_pub -h localhost -t "device/test/data" -m '{"device_id":"test"}'
```

### 问题4: 数据未上报到后端

**检查步骤**：
```bash
# 1. 检查MQTT Broker
docker logs dcim-emqx

# 2. 检查数据处理服务
docker logs dcim-data-processor

# 3. 验证数据流
mosquitto_sub -h localhost -t "dcim/collector/data" -v
```

---

## 附录

### A. 常用命令

```bash
# 查看Agent版本
./bin/collector-agent -version

# 验证配置文件
./bin/collector-agent -config config.yaml -validate

# 重载配置（需要支持）
kill -HUP $(pgrep collector-agent)

# 优雅停止
kill -TERM $(pgrep collector-agent)
```

### B. 日志级别

```yaml
# config.yaml
logging:
  level: "debug"  # debug/info/warn/error
  output: "file"  # file/stdout/both
  file: "logs/agent.log"
```

### C. 性能调优参数

```yaml
agent:
  max_concurrency: 1000      # 最大并发数
  heartbeat_interval: 30     # 心跳间隔

cache:
  max_cache_time: 24         # 缓存时长
  clean_interval: 10         # 清理间隔

receiver:
  mqtt_receiver:
    qos: 1                   # QoS级别
```

---

## 相关文档

- [双模式架构设计](./DUAL_MODE_ARCHITECTURE.md)
- [快速入门指南](./DUAL_MODE_QUICKSTART.md)
- [API文档](./API_REFERENCE.md)
- [协议扩展指南](./PROTOCOL_EXTENSION.md)

---

**技术支持**：如有问题，请提交Issue或联系维护团队。
