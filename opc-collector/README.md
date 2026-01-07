# OPC Collector - High-Scale Data Collection System

A high-performance OPC data collection system designed to handle 1.2 million OPC servers with 10-second collection intervals, writing to InfluxDB in batches.

## Features

- **Massive Scale**: Handles 1.2M OPC servers @ 10s intervals (~600K points/second)
- **Multi-Protocol**: Supports OPC UA, OPC DA (via gateway), and Gateway protocols
- **High Concurrency**: 2,000 concurrent connections per agent using worker pools
- **Batch Processing**: 10-second batch aggregation to InfluxDB
- **Resilience**: Local Badger DB cache for offline resilience
- **Fault Tolerance**: Per-device circuit breakers
- **Monitoring**: Prometheus metrics and Grafana dashboards

## Architecture

```
Configuration Files
        ↓
Collector Agents (600 instances) → InfluxDB Cluster
  - Connection Pools (UA/DA/Gateway)
  - Worker Pool (100 workers/agent)
  - 10s Batch Aggregator
  - Local Badger Cache
```

## Quick Start

### Prerequisites

- Go 1.21+
- InfluxDB 2.x
- Optional: Kubernetes cluster for deployment

### Build

```bash
make build
```

### Run

```bash
./bin/collector -config configs/config.yaml
```

## Configuration

See `configs/config.yaml.example` for full configuration options.

## Deployment

### Docker

```bash
docker build -f deploy/docker/Dockerfile.collector -t opc-collector:latest .
docker run -v $(pwd)/configs:/etc/opc-collector opc-collector:latest
```

### Kubernetes

```bash
kubectl apply -f deploy/k8s/
```

## Monitoring

Prometheus metrics are exposed on port 9090 by default.

## License

MIT
