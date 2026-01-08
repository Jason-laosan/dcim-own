# OPC Alert System

基于OPC Collector的实时数据处理和告警系统。

## 系统架构

```
OPC Collector (Go)
       ↓
   Kafka (opc-metrics topic)
       ↓
Alert Consumer (Spring Boot + Kafka)
       ↓
Data Processor (数据加工)
       ↓
InfluxDB (时序数据存储)
       ↓
Alert Engine (Flink 实时告警)
       ↓
PostgreSQL (告警规则配置)
       ↓
Alert Events (Kafka/通知)
```

## 模块说明

### 1. alert-common
通用数据模型和常量定义
- MetricData: Kafka消息模型
- ProcessedData: 加工后的数据模型
- AlertRule: 告警规则模型
- AlertTemplate: 告警模板模型
- AlertReceiver: 告警接收人模型
- AlertEvent: 告警事件模型

### 2. alert-consumer
Kafka消费者模块
- 从Kafka消费OPC采集的指标数据
- 调用数据加工处理服务
- 支持批量消费和单条消费

### 3. alert-processor
数据加工处理模块
- 数据验证
- 单位转换
- 计算字段
- 数据清洗
- 可扩展的数据转换逻辑

### 4. alert-storage
InfluxDB存储模块
- 写入时序数据到InfluxDB
- 支持批量写入
- Point数据转换
- 健康检查

### 5. alert-config
PostgreSQL配置管理模块
- 告警规则管理 (CRUD)
- 告警模板管理 (CRUD)
- 告警接收人管理 (CRUD)
- 配置缓存 (Spring Cache)
- 定时刷新配置

### 6. alert-engine
Flink实时告警引擎
- 从InfluxDB读取最近10分钟的数据
- 实时评估告警规则
- 支持连续违规次数判断
- 告警去重（5分钟内相同规则不重复告警）
- 发送告警事件到Kafka

## 技术栈

- **Java**: JDK 17
- **构建工具**: Maven
- **框架**: Spring Boot 3.2.1
- **实时计算**: Apache Flink 1.18.0
- **消息队列**: Apache Kafka 3.6.1
- **时序数据库**: InfluxDB 2.x
- **关系数据库**: PostgreSQL
- **连接池**: HikariCP 5.1.0
- **日志**: SLF4J + Logback
- **工具**: Lombok, Jackson

## 快速开始

### 1. 环境准备

确保已安装以下服务:
- PostgreSQL 12+
- InfluxDB 2.x
- Kafka 3.x
- JDK 17
- Maven 3.8+

### 2. 数据库初始化

```bash
# 创建PostgreSQL数据库
psql -U postgres -c "CREATE DATABASE opc_alert_db;"

# 初始化表结构
psql -U postgres -d opc_alert_db -f scripts/init-db.sql
```

### 3. 配置修改

修改配置文件中的数据库连接信息:

**alert-consumer/src/main/resources/application.properties**
```properties
influxdb.url=http://localhost:8086
influxdb.token=YOUR_INFLUXDB_TOKEN
influxdb.org=opc_organization
influxdb.bucket=opc_data

spring.kafka.bootstrap-servers=localhost:9092
```

**alert-engine/src/main/resources/application.properties**
```properties
spring.datasource.url=jdbc:postgresql://localhost:5432/opc_alert_db
spring.datasource.username=postgres
spring.datasource.password=YOUR_PASSWORD

influxdb.url=http://localhost:8086
influxdb.token=YOUR_INFLUXDB_TOKEN

kafka.bootstrap-servers=localhost:9092
```

### 4. 编译项目

```bash
cd opc-alert-system
mvn clean package -DskipTests
```

### 5. 启动服务

启动Consumer服务:
```bash
cd alert-consumer
mvn spring-boot:run
```

启动Alert Engine服务:
```bash
cd alert-engine
mvn spring-boot:run
```

## 告警规则配置

### 规则字段说明

- **rule_name**: 规则名称（唯一）
- **metric_name**: 监控的指标名称
- **condition_type**: 条件类型 (>, <, >=, <=, ==, !=)
- **threshold**: 阈值
- **level**: 告警级别 (INFO, WARNING, ERROR, CRITICAL)
- **time_window_seconds**: 时间窗口（秒）
- **consecutive_count**: 连续违规次数
- **device_filter**: 设备过滤（正则表达式）
- **template_id**: 告警模板ID

### 示例：添加告警规则

```sql
INSERT INTO alert_rules (
    rule_name, description, metric_name, condition_type, threshold,
    level, time_window_seconds, consecutive_count, template_id, enabled
)
VALUES (
    'High CPU Usage',
    'CPU使用率超过90%告警',
    'cpu_usage',
    '>',
    90.0,
    'WARNING',
    300,
    3,
    1,
    true
);
```

## 告警模板配置

模板支持变量替换:
- `${deviceId}`: 设备ID
- `${deviceIp}`: 设备IP
- `${metricName}`: 指标名称
- `${value}`: 当前值
- `${threshold}`: 阈值
- `${level}`: 告警级别
- `${timestamp}`: 时间戳

### 示例：添加告警模板

```sql
INSERT INTO alert_templates (
    template_name, title_template, message_template, channels, enabled
)
VALUES (
    'Custom Template',
    '[${level}] ${metricName} Alert',
    '设备 ${deviceId} 的 ${metricName} 当前值为 ${value}，超过阈值 ${threshold}',
    'EMAIL,SMS,WEBHOOK',
    true
);
```

## 告警接收人配置

### 示例：添加接收人

```sql
-- Email接收人
INSERT INTO alert_receivers (
    receiver_name, receiver_type, contact, level_filter, enabled
)
VALUES (
    'DevOps Team',
    'EMAIL',
    'devops@example.com',
    'ERROR,CRITICAL',
    true
);

-- Webhook接收人
INSERT INTO alert_receivers (
    receiver_name, receiver_type, contact, level_filter, enabled
)
VALUES (
    'Slack Webhook',
    'WEBHOOK',
    'https://hooks.slack.com/services/YOUR/WEBHOOK/URL',
    'CRITICAL',
    true
);
```

## 数据处理流程

1. **OPC Collector** 采集设备数据推送到Kafka (`opc-metrics` topic)
2. **Alert Consumer** 消费Kafka消息
3. **Data Processor** 加工处理数据（验证、转换、计算）
4. **InfluxDB Writer** 将处理后的数据写入InfluxDB
5. **Alert Engine** 从InfluxDB读取最近10分钟数据
6. **Flink Job** 实时评估告警规则
7. 触发告警则发送到Kafka (`alert-events` topic) 或直接通知

## 监控和运维

### 健康检查

Consumer服务: http://localhost:8081/actuator/health
Alert Engine服务: http://localhost:8082/actuator/health

### 日志查看

日志配置在 `application.properties` 中，可以调整日志级别:
```properties
logging.level.com.opc.alert=DEBUG
```

## 性能优化建议

1. **Kafka消费者**
   - 调整 `max-poll-records` 根据消息大小
   - 设置合适的 `concurrency` 并发数

2. **InfluxDB写入**
   - 使用批量写入
   - 合理设置时序数据保留策略

3. **Flink作业**
   - 调整并行度 `flink.parallelism`
   - 设置合适的检查点间隔

4. **数据库连接池**
   - HikariCP已优化配置
   - 根据负载调整连接池大小

## 扩展开发

### 自定义数据转换

在 `DataProcessorService.applyTransformations()` 方法中添加自定义逻辑。

### 自定义告警通知

实现新的通知渠道（Email, SMS, Webhook等），扩展告警发送逻辑。

### 添加新的告警规则类型

扩展 `AlertRule` 模型，支持更复杂的告警条件。

## License

MIT License

## 联系方式

如有问题，请提交Issue或联系开发团队。
