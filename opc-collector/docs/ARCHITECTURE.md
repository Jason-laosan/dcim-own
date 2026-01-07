# OPC Collector - Architecture Documentation

## Overview

The OPC Collector is a high-performance, scalable system designed to collect data from 1.2 million OPC servers with 10-second collection intervals, achieving approximately 600,000 messages per second throughput to Kafka.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                   OPC Servers (1.2M devices)                     │
│                    ↓  ↓  ↓  ↓  ↓  ↓  ↓  ↓                       │
└─────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│              Collector Agents (600 instances)                    │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Single Agent (handles ~2000 connections)                  │ │
│  │  ┌──────────────────────────────────────────────────────┐ │ │
│  │  │  Protocol Layer                                       │ │ │
│  │  │  - Connection Pools (OPC UA/DA/Gateway)             │ │ │
│  │  │  - 500-1000 persistent connections per protocol     │ │ │
│  │  └──────────────────────────────────────────────────────┘ │ │
│  │                         ↓                                   │ │
│  │  ┌──────────────────────────────────────────────────────┐ │ │
│  │  │  Scheduler                                            │ │ │
│  │  │  - 10-second collection intervals                    │ │ │
│  │  │  - Task management for all devices                   │ │ │
│  │  └──────────────────────────────────────────────────────┘ │ │
│  │                         ↓                                   │ │
│  │  ┌──────────────────────────────────────────────────────┐ │ │
│  │  │  Worker Pool (100 workers)                           │ │ │
│  │  │  - Semaphore-based concurrency control              │ │ │
│  │  │  - Parallel collection execution                     │ │ │
│  │  │  - Circuit breaker per device                        │ │ │
│  │  └──────────────────────────────────────────────────────┘ │ │
│  │                         ↓                                   │ │
│  │  ┌──────────────────────────────────────────────────────┐ │ │
│  │  │  Batch Aggregator                                     │ │ │
│  │  │  - 10-second time-based batching                     │ │ │
│  │  │  - Size-based flushing (10K points)                  │ │ │
│  │  │  - Memory-based flushing (256MB)                     │ │ │
│  │  └──────────────────────────────────────────────────────┘ │ │
│  │                         ↓                                   │ │
│  │  ┌──────────────────────────────────────────────────────┐ │ │
│  │  │  Local Cache (Badger DB)                             │ │ │
│  │  │  - Offline resilience                                │ │ │
│  │  │  - 24-hour TTL                                        │ │ │
│  │  │  - Background replay                                  │ │ │
│  │  └──────────────────────────────────────────────────────┘ │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                    Kafka Cluster (3-5 brokers)                   │
│  - Batch writes (JSON messages)                                 │
│  - ~600K messages/second total throughput                       │
│  - Configurable retention and partitioning                      │
└─────────────────────────────────────────────────────────────────┘
```

## Component Details

### 1. Protocol Layer

**Connection Pool** (`internal/protocol/connection_pool.go`)
- Manages persistent connections to OPC servers
- Prevents connection overhead for each collection
- Configurable max idle/open connections
- Automatic connection lifecycle management
- Connection health checking and cleanup

**OPC UA Protocol** (`internal/protocol/opcua.go`)
- Uses `gopcua` library for OPC UA communication
- Supports multiple security modes (None, Sign, SignAndEncrypt)
- Bulk read operations for efficiency
- Authentication support (username/password, certificates)
- Connection pooling integration

### 2. Scheduler

**Task Scheduler** (`internal/scheduler/scheduler.go`)
- Manages collection tasks for all devices
- 10-second interval scheduling per device
- Dynamic task addition/removal
- Task prioritization
- Automatic task submission to worker pool

### 3. Worker Pool

**Concurrent Execution** (`internal/collector/worker_pool.go`)
- Semaphore-based concurrency control (proven DCIM pattern)
- 100 workers per agent by default
- Task queue with backpressure handling
- Result and error channels
- Graceful shutdown support

**Circuit Breaker** (`internal/circuitbreaker/breaker.go`)
- Per-device fault isolation
- State machine: Closed → Open → Half-Open → Closed
- Configurable failure thresholds
- Automatic recovery attempts
- Prevents cascading failures

### 4. Batch Aggregator

**Batching Strategy** (`internal/batch/batcher.go`)
- Time-based: Flush every 10 seconds
- Size-based: Flush at 10,000 points
- Memory-based: Flush at 256MB buffer
- Dual-loop design: collection + periodic flush
- Memory-bounded buffer

### 5. Storage Layer

**Kafka Producer** (`internal/storage/kafka_producer.go`)
- JSON message serialization
- Retry logic with exponential backoff
- Snappy compression
- Partitioning by device ID
- Error handling with cache fallback

**Local Cache** (`internal/cache/cache.go`)
- Badger DB for persistence
- Write-ahead logging for failed writes
- 24-hour TTL
- Background replay to Kafka on recovery
- Automatic garbage collection

### 6. Monitoring

**Prometheus Metrics** (`internal/monitor/metrics.go`)
- Collection metrics (count, duration, errors)
- Worker pool statistics
- Connection pool utilization
- Batch processing metrics
- Kafka write performance
- Circuit breaker states
- Memory and CPU metrics

## Data Flow

1. **Scheduling**
   - Scheduler checks tasks every second
   - Tasks with `NextRun` ≤ now are submitted to worker pool

2. **Collection**
   - Worker acquires semaphore slot
   - Gets connection from pool
   - Reads metrics from OPC server
   - Releases connection back to pool
   - Sends result to result channel

3. **Batching**
   - Batcher receives results from worker pool
   - Aggregates in memory buffer
   - Flushes on timer (10s) or size/memory threshold

4. **Storage**
   - Converts metrics to JSON format
   - Sends messages to Kafka topic
   - Retries on failure (up to 3 times)
   - Falls back to local cache on persistent failure

5. **Recovery**
   - Background process monitors cache
   - Replays cached data when Kafka recovers
   - Automatic cleanup after successful replay

## Scalability

### Vertical Scaling (Per Agent)
- **CPU**: 8-16 cores recommended
- **Memory**: 16-32 GB recommended
- **Concurrency**: Up to 200 workers
- **Connections**: Up to 2,500 devices per agent

### Horizontal Scaling (Agent Count)
- **Formula**: Agents = Total Devices / Devices per Agent
- **Example**: 1,200,000 / 2,000 = 600 agents
- **Distribution**: Geographic or protocol-based
- **Coordination**: Stateless agents, no coordination needed

### Kafka Scaling
- **Write Rate**: ~600,000 messages/second
- **Cluster**: 3-5 brokers recommended
- **Storage**: 2-4 TB NVMe SSD per broker
- **Retention**: Configurable (7-30 days typical)
- **Partitions**: 50-100 partitions for parallelism

## Performance Characteristics

### Throughput
- **Per Agent**: ~1,000 collections/second (10s interval, 2K devices)
- **Total System**: ~600,000 messages/second (600 agents × 1K messages/s)
- **Kafka Write**: JSON messages with snappy compression

### Latency
- **Collection**: < 100ms average (OPC UA bulk read)
- **Batch Flush**: < 5 seconds (configurable)
- **End-to-End**: < 15 seconds (collection + batching + write)

### Resource Usage
- **CPU**: 40-60% under normal load (8 cores)
- **Memory**: 8-16 GB steady state (with GC tuning)
- **Network**: ~10 Mbps per agent (depends on metric count)
- **Disk I/O**: Minimal (cache writes only on Kafka failure)

## Reliability

### Fault Tolerance
- **Circuit Breakers**: Isolate unhealthy devices
- **Connection Pooling**: Automatic reconnection on failure
- **Retry Logic**: Exponential backoff for transient failures
- **Local Cache**: Data preservation during Kafka downtime

### High Availability
- **Stateless Agents**: No single point of failure
- **Kubernetes**: Auto-restart on crash
- **Rolling Updates**: Zero-downtime deployments
- **Graceful Shutdown**: Clean resource cleanup

### Data Integrity
- **At-Most-Once**: No duplicate writes
- **Eventual Consistency**: Cache replay ensures data delivery
- **Quality Tracking**: OPC quality codes preserved
- **Timestamps**: Server timestamps maintained

## Configuration Tuning

### High Throughput
```yaml
agent:
  max_concurrency: 150          # More workers
  gc_percent: 50                # Aggressive GC

connection_pool:
  opcua:
    max_open: 1000              # More connections

batch:
  max_size: 20000               # Larger batches
  max_memory_mb: 512            # More buffering
```

### Low Latency
```yaml
batch:
  interval: 5                   # Faster flushing
  max_size: 5000                # Smaller batches

kafka:
  timeout: 10                   # Faster timeout
  compression: "lz4"            # Faster compression
```

### Memory Constrained
```yaml
agent:
  max_concurrency: 50           # Fewer workers
  gc_percent: 30                # More aggressive GC

batch:
  max_size: 5000                # Smaller batches
  max_memory_mb: 128            # Less buffering

connection_pool:
  opcua:
    max_open: 300               # Fewer connections
```

## Security

### OPC UA Security
- TLS encryption support
- Certificate-based authentication
- Username/password authentication
- Security policy configuration

### Deployment Security
- Non-root container user
- Read-only filesystem (except cache)
- Network policies (Kubernetes)
- Secret management (Kubernetes Secrets)

### Data Security
- Encrypted Kafka connections (TLS)
- SASL authentication support
- RBAC for Kubernetes resources

## Monitoring & Observability

### Metrics
- Prometheus exposition format
- 30+ custom metrics
- Standard Go runtime metrics
- HTTP endpoint: `http://agent:9090/metrics`

### Logging
- Structured JSON logging (zap)
- Configurable log levels
- Log rotation (lumberjack)
- Contextual logging (device ID, agent ID, etc.)

### Health Checks
- Liveness probe: Metrics endpoint availability
- Readiness probe: Kafka connectivity
- Startup probe: Initial device load

## Future Enhancements

### Planned Features
1. OPC DA protocol support (via gateway)
2. Gateway aggregator protocol
3. Database device source (PostgreSQL/MySQL)
4. etcd device source for dynamic configuration
5. mTLS for inter-service communication
6. Distributed tracing (Jaeger/Zipkin)
7. Advanced circuit breaker strategies
8. Auto-scaling based on queue depth
9. Real-time anomaly detection
10. Multi-destination routing (multiple Kafka topics, etc.)

### Performance Improvements
1. Connection multiplexing
2. Subscription-based collection (OPC UA subscriptions)
3. Compression for cache storage
4. Zero-copy serialization
5. SIMD optimizations for line protocol generation
