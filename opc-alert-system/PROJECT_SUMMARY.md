# OPC Alert System - 项目总结

## 项目概述

基于opc-collector项目的实时数据处理和告警系统，使用Java + Maven构建，集成了Spring Boot、Apache Flink、InfluxDB、Kafka和PostgreSQL。

## 技术栈

- **语言**: Java 17
- **构建工具**: Maven
- **核心框架**: Spring Boot 3.2.1
- **实时计算**: Apache Flink 1.18.0
- **消息队列**: Apache Kafka 3.6.1
- **时序数据库**: InfluxDB 2.x
- **关系数据库**: PostgreSQL
- **连接池**: HikariCP 5.1.0

## 模块架构

### 1. alert-common (通用模块)
**功能**: 提供所有模块共享的数据模型和常量
**主要类**:
- `MetricData`: OPC采集数据模型（来自Kafka）
- `ProcessedData`: 加工后的数据模型
- `AlertRule`: 告警规则配置
- `AlertTemplate`: 告警消息模板
- `AlertReceiver`: 告警接收人
- `AlertEvent`: 告警事件
- `Constants`: 常量定义

### 2. alert-consumer (Kafka消费者)
**功能**: 消费OPC数据并触发数据处理
**核心组件**:
- `KafkaConsumerConfig`: Kafka消费者配置
- `MetricConsumerService`: 消息消费服务
- `AlertConsumerApplication`: 启动类

**特性**:
- 批量消费（默认500条/批次）
- 并发消费（默认3个并发）
- 手动提交偏移量
- 错误处理和重试

### 3. alert-processor (数据加工处理)
**功能**: 对原始数据进行转换和加工
**核心组件**:
- `DataProcessorService`: 数据处理服务

**处理功能**:
- 数据验证
- 单位转换（示例：摄氏度转华氏度）
- 计算字段
- 数据清洗
- 时间戳标签化

### 4. alert-storage (InfluxDB存储)
**功能**: 将处理后的数据写入InfluxDB
**核心组件**:
- `InfluxDBConfig`: InfluxDB配置
- `InfluxDBWriter`: 数据写入服务

**特性**:
- 单条写入和批量写入
- Point数据转换
- 自动类型处理
- 健康检查

### 5. alert-config (PostgreSQL配置管理)
**功能**: 管理告警规则、模板和接收人
**核心组件**:
- `DataSourceConfig`: HikariCP数据源配置
- `AlertRuleRepository`: 告警规则仓储
- `AlertTemplateRepository`: 模板仓储
- `AlertReceiverRepository`: 接收人仓储
- `AlertConfigService`: 配置服务（带缓存）

**特性**:
- CRUD操作
- Spring Cache缓存
- 定时刷新（每60秒）
- 模板变量替换

### 6. alert-engine (Flink告警引擎)
**功能**: 实时评估告警规则并触发告警
**核心组件**:
- `FlinkAlertJob`: Flink作业主类
- `AlertEvaluationFunction`: 告警评估函数
- `InfluxDBSource`: InfluxDB数据源
- `AlertEngineApplication`: 启动类

**特性**:
- 从InfluxDB读取最近10分钟数据
- 基于设备ID的KeyedStream
- 状态管理（违规计数、去重）
- 连续违规判断
- 告警去重（5分钟内相同规则不重复）
- 检查点机制
- 发送告警到Kafka

## 数据流程

```
1. OPC Collector采集数据 → Kafka (opc-metrics)
2. Alert Consumer消费 → Data Processor加工
3. 加工后数据 → InfluxDB存储
4. Flink从InfluxDB读取最近10分钟数据
5. Alert Engine实时评估规则
6. 触发告警 → Kafka (alert-events)
7. 通知系统 → 告警接收人
```

## 数据库设计

### PostgreSQL表结构

**alert_rules** - 告警规则表
- 支持多种条件类型（>, <, >=, <=, ==, !=）
- 告警级别（INFO, WARNING, ERROR, CRITICAL）
- 时间窗口和连续次数判断
- 设备过滤（正则表达式）

**alert_templates** - 告警模板表
- 支持变量替换
- 多通道配置（EMAIL, SMS, WEBHOOK）

**alert_receivers** - 接收人表
- 多种接收类型
- 告警级别过滤

**alert_history** - 告警历史表（可选）
- 记录所有触发的告警

## 配置说明

### alert-consumer配置
- Kafka消费者配置
- InfluxDB连接配置
- 日志配置

### alert-engine配置
- PostgreSQL连接配置
- InfluxDB连接配置
- Kafka生产者配置
- Flink作业配置
- 缓存和调度配置

## Docker Compose环境

包含所有基础设施服务:
- PostgreSQL (端口: 5432)
- InfluxDB (端口: 8086)
- Kafka + Zookeeper (端口: 9092, 2181)
- Kafka UI (端口: 8080)

## 快速启动

### Windows
```cmd
start.bat
```

### Linux/Mac
```bash
chmod +x start.sh
./start.sh
```

然后分别启动两个服务:
```bash
# Terminal 1
cd alert-consumer
mvn spring-boot:run

# Terminal 2
cd alert-engine
mvn spring-boot:run
```

## 告警规则示例

### 高温告警
```sql
INSERT INTO alert_rules (
    rule_name, metric_name, condition_type, threshold,
    level, time_window_seconds, consecutive_count, template_id
) VALUES (
    'High Temperature', 'temperature', '>', 80.0,
    'WARNING', 600, 3, 1
);
```

### 临界压力告警
```sql
INSERT INTO alert_rules (
    rule_name, metric_name, condition_type, threshold,
    level, time_window_seconds, consecutive_count, template_id
) VALUES (
    'Critical Pressure', 'pressure', '>', 100.0,
    'CRITICAL', 300, 2, 1
);
```

## 扩展功能

### 可扩展点
1. **数据处理**: `DataProcessorService.applyTransformations()`
2. **告警通知**: 实现新的通知渠道
3. **告警规则**: 支持更复杂的条件表达式
4. **数据源**: 支持从其他数据源读取

### 性能优化
- Kafka批量消费
- InfluxDB批量写入
- PostgreSQL连接池优化
- Flink检查点和状态管理
- Spring Cache缓存配置

## 监控和运维

### 健康检查
- Consumer: http://localhost:8081/actuator/health
- Engine: http://localhost:8082/actuator/health

### Kafka监控
- Kafka UI: http://localhost:8080

### InfluxDB管理
- UI: http://localhost:8086

### 日志
所有服务都配置了详细的日志输出，可在控制台查看。

## 注意事项

1. **首次启动**: 确保先启动Docker Compose中的基础设施
2. **配置修改**: 根据实际环境修改`application.properties`
3. **InfluxDB Token**: 使用Docker Compose时token为`my-super-secret-auth-token`
4. **Flink状态**: Flink状态持久化需配置检查点存储
5. **生产环境**: 建议配置外部配置中心（如Spring Cloud Config）

## 项目文件清单

```
opc-alert-system/
├── pom.xml                          # 父POM
├── README.md                        # 项目说明
├── docker-compose.yml               # Docker编排
├── start.sh / start.bat             # 启动脚本
├── scripts/
│   └── init-db.sql                  # 数据库初始化
├── alert-common/                    # 通用模块
├── alert-consumer/                  # Kafka消费者
├── alert-processor/                 # 数据处理
├── alert-storage/                   # InfluxDB存储
├── alert-config/                    # PostgreSQL配置
└── alert-engine/                    # Flink告警引擎
```

## 总结

本项目实现了一个完整的实时数据处理和告警系统，具有以下特点:

✅ 模块化设计，职责清晰
✅ 支持动态配置告警规则
✅ 实时计算和告警（10分钟时间窗口）
✅ 告警去重和连续判断
✅ 灵活的模板系统
✅ 可扩展的数据处理
✅ 完整的文档和示例
✅ Docker一键部署开发环境

适用于工业设备监控、IoT数据处理、实时告警等场景。
