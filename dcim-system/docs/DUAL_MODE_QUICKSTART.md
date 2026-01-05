# 双模式采集快速入门指南

## 快速开始

### 1. 选择配置模式

根据您的数据中心设备情况，选择合适的配置文件：

```bash
# 双模式（推荐）- 同时支持主动拉取和被动接收
cp config-dual-mode.yaml.example config.yaml

# 仅主动拉取 - 传统设备
cp config-pull-only.yaml.example config.yaml

# 仅被动接收 - 智能设备
cp config-push-only.yaml.example config.yaml
```

### 2. 修改配置

编辑 `config.yaml`，修改以下关键配置：

```yaml
agent:
  id: "your-agent-id"           # 修改为您的Agent ID
  enable_pull_mode: true         # 是否启用主动拉取
  enable_push_mode: true         # 是否启用被动接收

mqtt:
  broker: "tcp://your-mqtt:1883" # 修改为您的MQTT地址
  username: "your-username"
  password: "your-password"

receiver:
  mqtt_receiver:
    broker: "tcp://your-mqtt:1883"
    subscribe_topics:
      - "device/+/data"          # 修改为您的Topic
```

### 3. 启动Agent

```bash
# 编译
go build -o bin/collector-agent cmd/main.go

# 运行
./bin/collector-agent -config config.yaml
```

### 4. 验证运行

查看日志确认两种模式是否正常启动：

```
[INFO] agent started successfully
[INFO] pull mode started          # 主动拉取模式已启动
[INFO] push mode started          # 被动接收模式已启动
[INFO] MQTT receiver connected    # MQTT接收器已连接
```

---

## 使用示例

### 示例1：添加Pull模式采集任务（SNMP）

通过采集管理服务API添加任务：

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-001",
    "task_id": "task-snmp-001",
    "device_id": "switch-001",
    "device_ip": "192.168.1.100",
    "device_type": "switch",
    "protocol": "snmp",
    "mode": "pull",
    "interval": 30,
    "metrics": [
      "1.3.6.1.2.1.1.3.0",
      "1.3.6.1.2.1.2.2.1.10"
    ],
    "config": {
      "version": "v2c",
      "community": "public"
    }
  }'
```

### 示例2：设备通过MQTT推送数据（Push模式）

智能设备向Agent推送数据：

```python
import paho.mqtt.client as mqtt
import json
import time

# 连接MQTT
client = mqtt.Client()
client.username_pw_set("device_user", "device_password")
client.connect("localhost", 1883)

# 推送数据
data = {
    "device_id": "smart-pdu-001",
    "device_ip": "192.168.1.200",
    "device_type": "pdu",
    "timestamp": time.time(),
    "metrics": {
        "voltage": 220.5,
        "current": 15.2,
        "power": 3350.0
    },
    "status": "success"
}

client.publish("device/smart-pdu-001/data", json.dumps(data))
```

### 示例3：Modbus设备推送数据（Push模式）

Agent作为Modbus Slave，接收Modbus主站推送的数据：

```python
from pymodbus.client import ModbusTcpClient

# 连接到Agent的Modbus接收器
client = ModbusTcpClient('localhost', port=502)

# 写入寄存器（推送数据）
client.write_registers(0, [220, 152, 3350], unit=1)
```

---

## 常见问题

### Q1: 如何同时使用两种模式？

**A**: 在配置文件中同时启用两种模式：

```yaml
agent:
  enable_pull_mode: true
  enable_push_mode: true

receiver:
  enabled: true
```

### Q2: Pull模式支持哪些协议？

**A**: 支持以下协议：
- SNMP (v1/v2c/v3)
- IPMI
- Modbus (TCP/RTU)
- HTTP/RESTful
- SSH/Telnet
- BACnet/IP

### Q3: Push模式支持哪些协议？

**A**: 支持以下协议：
- MQTT订阅
- Modbus Slave模式

### Q4: 如何查看Agent状态？

**A**: Agent会定期发送心跳到MQTT Topic：

```bash
# 订阅心跳Topic
mosquitto_sub -h localhost -t "dcim/collector/data/heartbeat"
```

### Q5: 数据格式是什么？

**A**: 统一的JSON格式：

```json
{
  "device_id": "device-001",
  "device_ip": "192.168.1.100",
  "device_type": "switch",
  "timestamp": "2026-01-05T13:55:00Z",
  "metrics": {
    "cpu_usage": 45.2,
    "memory_usage": 60.5
  },
  "status": "success",
  "error": ""
}
```

---

## 性能调优

### Pull模式优化

```yaml
agent:
  max_concurrency: 1000  # 最大并发采集数，根据机器性能调整

# 任务配置
interval: 30  # 采集间隔（秒），根据实际需求调整
```

### Push模式优化

```yaml
receiver:
  mqtt_receiver:
    qos: 1  # QoS级别：0（最多一次）、1（至少一次）、2（恰好一次）
```

### 缓存优化

```yaml
cache:
  max_cache_time: 24  # 最大缓存时长（小时）
  clean_interval: 10  # 清理间隔（分钟）
```

---

## 故障排查

### 问题1: Pull模式采集失败

**检查**：
1. 设备IP是否可达：`ping 192.168.1.100`
2. 协议端口是否开放：`telnet 192.168.1.100 161`（SNMP）
3. 查看Agent日志：`grep "collect failed" logs/agent.log`

### 问题2: Push模式未接收到数据

**检查**：
1. MQTT Broker是否正常：`mosquitto_sub -h localhost -t "#"`
2. Topic是否匹配：确认设备推送的Topic在订阅列表中
3. 数据格式是否正确：确认JSON格式符合DeviceData结构

### 问题3: Agent频繁重连

**检查**：
1. 网络是否稳定
2. MQTT Broker是否过载
3. 调整重连参数：`connect_retry_interval: 30`

---

## 下一步

- 查看完整架构设计：[DUAL_MODE_ARCHITECTURE.md](./DUAL_MODE_ARCHITECTURE.md)
- 添加自定义协议插件
- 配置告警规则
- 部署到生产环境

---

**技术支持**：如有问题，请提交Issue或联系维护团队。
