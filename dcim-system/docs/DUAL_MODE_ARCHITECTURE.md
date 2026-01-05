# DCIM采集Agent双模式架构设计

## 一、概述

本文档描述DCIM采集Agent支持**主动拉取（Pull）**和**被动接收（Push）**两种数据采集模式的架构设计。

### 设计目标

1. **灵活性**：支持两种采集模式，可独立启用或同时启用
2. **扩展性**：协议插件化，易于扩展新协议
3. **高性能**：支持高并发采集和数据接收
4. **可靠性**：断线重连、本地缓存、数据补发

---

## 二、采集模式对比

### 2.1 主动拉取模式（Pull Mode）

**工作原理**：Agent按照配置的时间间隔，主动向设备发起采集请求。

**支持协议**：
- ✅ SNMP (v1/v2c/v3) - 网络设备、PDU、UPS、空调
- ✅ IPMI - 物理服务器
- ✅ Modbus (TCP/RTU) - 温湿度传感器、智能电表
- ✅ HTTP/RESTful - 智能PDU、云化设备
- ✅ SSH/Telnet - 服务器/网络设备命令行
- ✅ BACnet/IP - 楼宇自控、空调系统

**适用场景**：
- 传统设备（无主动上报能力）
- 需要精确控制采集频率
- 设备数量较多，需要统一调度

**优点**：
- 采集时机可控
- 支持协议广泛
- 适配传统设备

**缺点**：
- 轮询开销较大
- 实时性相对较低
- 需要维护设备连接

---

### 2.2 被动接收模式（Push Mode）

**工作原理**：Agent被动监听，接收设备主动推送的数据。

**支持协议**：
- ✅ MQTT订阅 - 智能设备、IoT传感器
- ✅ Modbus Slave - Modbus主站推送数据

**适用场景**：
- 智能设备（支持主动上报）
- 事件驱动型数据采集
- 实时性要求高的场景

**优点**：
- 实时性高（事件驱动）
- 无轮询开销
- 设备主动推送，减少Agent负载

**缺点**：
- 依赖设备支持主动推送
- 协议支持相对有限
- 需要设备配置推送地址

---

## 三、架构设计

### 3.1 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      DCIM Agent                              │
│                                                               │
│  ┌──────────────────┐          ┌──────────────────┐         │
│  │  主动拉取模式     │          │  被动接收模式     │         │
│  │  (Pull Mode)     │          │  (Push Mode)     │         │
│  └──────────────────┘          └──────────────────┘         │
│           │                              │                   │
│  ┌────────▼────────┐          ┌─────────▼─────────┐        │
│  │   Scheduler     │          │    Receiver       │        │
│  │   任务调度器     │          │    接收器管理器    │        │
│  └────────┬────────┘          └─────────┬─────────┘        │
│           │                              │                   │
│           │    ┌─────────────────┐      │                   │
│           └───►│   Collector     │◄─────┘                   │
│                │   数据采集器     │                          │
│                └────────┬────────┘                          │
│                         │                                    │
│                ┌────────▼────────┐                          │
│                │  Protocol Pool  │                          │
│                │  协议插件池      │                          │
│                └────────┬────────┘                          │
│                         │                                    │
│           ┌─────────────┼─────────────┐                    │
│           │             │             │                     │
│      ┌────▼───┐   ┌────▼───┐   ┌────▼───┐                │
│      │ SNMP   │   │ MQTT   │   │Modbus  │   ...          │
│      │Protocol│   │Protocol│   │Protocol│                 │
│      └────────┘   └────────┘   └────────┘                 │
│                                                              │
│  ┌──────────────────────────────────────────────┐          │
│  │           Data Pipeline (统一数据流转)        │          │
│  │  Cache → Publish → MQTT Broker → Backend    │          │
│  └──────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 核心组件

#### 3.2.1 Agent（代理核心）

**职责**：
- 管理采集器、调度器、接收器的生命周期
- 统一数据上报通道
- 心跳上报、缓存数据重发

**关键代码**：`collector-agent/internal/agent/agent.go`

#### 3.2.2 Scheduler（任务调度器）- Pull模式

**职责**：
- 管理采集任务（添加、删除、更新）
- 按时间间隔或Cron表达式触发采集
- 控制并发采集数量

**关键代码**：`collector-agent/internal/scheduler/scheduler.go`

#### 3.2.3 Receiver（接收器管理器）- Push模式

**职责**：
- 管理MQTT接收器和Modbus接收器
- 接收设备推送的数据
- 数据格式转换和校验

**关键代码**：`collector-agent/internal/receiver/receiver.go`

**子组件**：
- **MQTTReceiver**：订阅MQTT Topic，接收设备推送的JSON数据
- **ModbusReceiver**：作为Modbus Slave，接收Modbus主站推送的数据

#### 3.2.4 Collector（数据采集器）

**职责**：
- 管理协议插件池
- 执行具体的数据采集
- 批量采集并发控制

**关键代码**：`collector-agent/internal/collector/collector.go`

#### 3.2.5 Protocol（协议插件）

**接口定义**：
```go
type Protocol interface {
    Name() string
    Collect(ctx context.Context, task *CollectTask) (*DeviceData, error)
    Validate(config map[string]interface{}) error
    SupportedModes() []CollectMode  // 返回支持的采集模式
    Close() error
}
```

**协议模式支持**：
- SNMP：仅支持Pull模式
- MQTT：支持Push模式（订阅）和Pull模式（请求/响应）
- Modbus：支持Pull模式（主站）和Push模式（从站）
- HTTP：仅支持Pull模式

---

## 四、数据流转

### 4.1 主动拉取模式数据流

```
1. Scheduler触发采集任务
   ↓
2. Collector调用协议插件执行采集
   ↓
3. Protocol连接设备，读取数据
   ↓
4. 数据返回给Collector，进行预处理
   ↓
5. Agent发布数据到MQTT Broker
   ↓
6. 数据处理服务消费数据，写入数据库
```

### 4.2 被动接收模式数据流

```
1. 设备主动推送数据到Agent
   ↓
2. Receiver接收数据（MQTT订阅 或 Modbus Slave）
   ↓
3. 数据格式解析和校验
   ↓
4. 调用Agent的数据处理回调
   ↓
5. Agent发布数据到MQTT Broker
   ↓
6. 数据处理服务消费数据，写入数据库
```

### 4.3 统一数据模型

无论是Pull还是Push模式，数据最终都转换为统一的`DeviceData`结构：

```go
type DeviceData struct {
    DeviceID   string                 // 设备ID
    DeviceIP   string                 // 设备IP
    DeviceType string                 // 设备类型
    Timestamp  time.Time              // 采集时间
    Metrics    map[string]interface{} // 采集指标
    Status     string                 // 采集状态
    Error      string                 // 错误信息
}
```

---

## 五、配置方案

### 5.1 双模式配置（推荐）

适用于混合场景，既有传统设备又有智能设备。

```yaml
agent:
  enable_pull_mode: true   # 启用主动拉取
  enable_push_mode: true   # 启用被动接收

receiver:
  enabled: true
  mqtt_receiver:
    enabled: true
    subscribe_topics:
      - "device/+/data"
  modbus_receiver:
    enabled: true
    mode: "tcp"
    listen_addr: "0.0.0.0:502"
```

**配置文件**：`config-dual-mode.yaml.example`

### 5.2 仅主动拉取配置

适用于传统数据中心，设备无主动上报能力。

```yaml
agent:
  enable_pull_mode: true
  enable_push_mode: false

receiver:
  enabled: false
```

**配置文件**：`config-pull-only.yaml.example`

### 5.3 仅被动接收配置

适用于智能化数据中心，设备支持主动推送。

```yaml
agent:
  enable_pull_mode: false
  enable_push_mode: true

receiver:
  enabled: true
  mqtt_receiver:
    enabled: true
  modbus_receiver:
    enabled: true
```

**配置文件**：`config-push-only.yaml.example`

---

## 六、协议扩展指南

### 6.1 添加支持Pull模式的协议

1. 在`internal/protocol/`目录创建协议实现文件
2. 实现`Protocol`接口
3. 在`SupportedModes()`方法中返回`[]CollectMode{CollectModePull}`
4. 在Agent中注册协议

示例：
```go
type HTTPProtocol struct {
    client *http.Client
}

func (h *HTTPProtocol) SupportedModes() []CollectMode {
    return []CollectMode{CollectModePull}
}

func (h *HTTPProtocol) Collect(ctx context.Context, task *CollectTask) (*DeviceData, error) {
    // 实现HTTP采集逻辑
}
```

### 6.2 添加支持Push模式的协议

1. 在`internal/receiver/`目录创建接收器实现
2. 实现数据监听和接收逻辑
3. 将接收到的数据转换为`DeviceData`
4. 调用数据处理回调

示例：
```go
type WebSocketReceiver struct {
    server *websocket.Server
    handler DataHandler
}

func (w *WebSocketReceiver) Start() error {
    // 启动WebSocket服务器
    // 接收数据后调用 w.handler(data)
}
```

---

## 七、性能优化

### 7.1 Pull模式优化

- **并发控制**：使用协程池限制最大并发数（默认1000）
- **批量采集**：支持批量下发任务，减少调度开销
- **连接复用**：协议插件支持连接池，减少连接建立开销
- **智能调度**：支持Cron表达式，灵活控制采集时机

### 7.2 Push模式优化

- **异步处理**：接收器使用异步回调，不阻塞数据接收
- **缓冲队列**：使用带缓冲的通道，削峰填谷
- **订阅过滤**：MQTT使用通配符订阅，减少订阅数量
- **数据校验**：接收端进行数据格式校验，过滤无效数据

### 7.3 通用优化

- **本地缓存**：使用badger KV存储，断网时缓存数据
- **断线重连**：MQTT自动重连，指数退避策略
- **数据压缩**：大批量数据支持压缩传输
- **心跳优化**：心跳间隔可配置，减少网络开销

---

## 八、使用场景示例

### 场景1：混合数据中心

**需求**：
- 传统网络设备（交换机、路由器）：使用SNMP采集
- 智能PDU：支持MQTT主动推送
- 温湿度传感器：Modbus RTU采集

**方案**：
```yaml
agent:
  enable_pull_mode: true   # Pull模式采集传统设备
  enable_push_mode: true   # Push模式接收智能设备

receiver:
  mqtt_receiver:
    enabled: true
    subscribe_topics:
      - "smart_pdu/+/metrics"  # 接收智能PDU推送
```

**任务配置**：
```json
// Pull模式任务 - SNMP采集交换机
{
  "task_id": "task-001",
  "device_id": "switch-001",
  "protocol": "snmp",
  "mode": "pull",
  "interval": 30
}

// Push模式 - 智能PDU自动推送，无需配置任务
```

### 场景2：纯智能化数据中心

**需求**：
- 所有设备支持MQTT主动推送
- 实时性要求高

**方案**：
```yaml
agent:
  enable_pull_mode: false  # 禁用Pull模式
  enable_push_mode: true   # 仅启用Push模式

receiver:
  mqtt_receiver:
    enabled: true
    subscribe_topics:
      - "device/+/data"
      - "sensor/+/metrics"
```

### 场景3：传统数据中心

**需求**：
- 设备无主动上报能力
- 需要定时采集

**方案**：
```yaml
agent:
  enable_pull_mode: true   # 仅启用Pull模式
  enable_push_mode: false

receiver:
  enabled: false
```

---

## 九、监控与告警

### 9.1 Agent状态监控

- **心跳上报**：每30秒上报Agent状态
- **模式状态**：上报当前启用的采集模式
- **任务统计**：Pull模式任务数量、执行情况
- **接收统计**：Push模式接收数据量、错误率

### 9.2 告警规则

- Agent心跳超时（>90秒）
- Pull模式采集失败率 > 10%
- Push模式数据接收中断
- 本地缓存数据积压 > 1000条

---

## 十、最佳实践

### 10.1 模式选择建议

| 设备类型 | 推荐模式 | 理由 |
|---------|---------|------|
| 传统网络设备 | Pull | 无主动推送能力 |
| 智能PDU | Push | 支持MQTT推送，实时性高 |
| 温湿度传感器 | Pull | Modbus采集稳定 |
| 云化设备 | Pull (HTTP) | RESTful API采集 |
| IoT传感器 | Push (MQTT) | 事件驱动，省电 |

### 10.2 配置建议

1. **混合场景优先使用双模式**，灵活适配不同设备
2. **Pull模式采集间隔**：根据设备类型调整（10-300秒）
3. **Push模式Topic设计**：使用层级结构，便于过滤
4. **并发数设置**：根据Agent部署机器性能调整（500-2000）
5. **缓存时长**：建议24小时，避免数据丢失

### 10.3 安全建议

- MQTT通信启用TLS加密
- 使用强密码认证
- 限制订阅Topic权限
- 定期更新协议插件
- 监控异常数据推送

---

## 十一、总结

本架构设计实现了**主动拉取**和**被动接收**两种数据采集模式的统一管理，具有以下优势：

✅ **灵活性**：支持两种模式独立或混合使用  
✅ **扩展性**：协议插件化，易于扩展  
✅ **高性能**：支持高并发采集和数据接收  
✅ **可靠性**：断线重连、本地缓存、数据补发  
✅ **易用性**：配置驱动，无需修改代码  

适用于各类数据中心场景，从传统数据中心到智能化数据中心均可灵活适配。
