# Quick Start Guide - OPC Collector

This guide will help you get the OPC Collector up and running quickly.

## Prerequisites

- Go 1.21 or later
- InfluxDB 2.x instance
- OPC UA servers to collect from

## Step 1: Configuration

1. Copy the example configuration files:
```bash
cp configs/config.yaml.example configs/config.yaml
cp configs/machines.yaml.example configs/machines.yaml
```

2. Edit `configs/config.yaml`:
   - Update `influxdb.url` with your InfluxDB URL
   - Update `influxdb.token` with your InfluxDB token
   - Update `influxdb.org` and `influxdb.bucket` as needed
   - Adjust `agent.max_concurrency` based on your hardware (default: 100)
   - Adjust `agent.max_devices` based on device count (default: 2000)

3. Edit `configs/machines.yaml`:
   - Add your OPC UA server endpoints
   - Configure metrics to collect
   - Set collection intervals

Example device configuration:
```yaml
devices:
  - id: "opc-server-001"
    name: "Temperature Sensor"
    ip: "192.168.1.100"
    port: 4840
    protocol: "opcua"
    enabled: true
    connection_config:
      security_mode: "None"
      timeout: 10
    metrics:
      - node_id: "ns=2;s=Temperature"
        name: "temperature"
        data_type: "float"
        unit: "celsius"
    interval: 10
    tags:
      location: "warehouse-1"
```

## Step 2: Build

```bash
make build
```

Or manually:
```bash
go build -o bin/collector ./cmd/collector
```

## Step 3: Run

```bash
./bin/collector -config configs/config.yaml
```

Or using make:
```bash
make run
```

## Step 4: Verify

1. Check that the collector is running:
```bash
# Check logs for startup messages
# Look for "OPC Collector running" message
```

2. Access Prometheus metrics:
```bash
curl http://localhost:9090/metrics
```

3. Verify data in InfluxDB:
```bash
# Using InfluxDB CLI
influx query 'from(bucket:"opc-metrics") |> range(start: -5m)'
```

## Monitoring

The collector exposes Prometheus metrics on port 9090 by default.

Key metrics to monitor:
- `opc_collection_total` - Total collections
- `opc_collection_duration_seconds` - Collection latency
- `opc_worker_pool_active_workers` - Active workers
- `opc_batch_flush_duration_seconds` - Batch write performance
- `opc_influxdb_points_written_total` - Points written to InfluxDB

## Docker Deployment

Build the Docker image:
```bash
make docker-build
```

Run with Docker:
```bash
docker run -v $(pwd)/configs:/etc/opc-collector opc-collector:latest
```

## Kubernetes Deployment

1. Create the namespace and resources:
```bash
kubectl apply -f deploy/k8s/collector-statefulset.yaml
```

2. Create machines ConfigMap:
```bash
kubectl create configmap opc-collector-machines \
  --from-file=machines.yaml=configs/machines.yaml \
  -n opc-system
```

3. Scale the StatefulSet:
```bash
kubectl scale statefulset opc-collector --replicas=10 -n opc-system
```

## Scaling for 1.2M Devices

To handle 1.2 million devices:

1. Calculate required agents:
   - Each agent handles ~2,000 devices
   - Required agents: 1,200,000 / 2,000 = 600 agents

2. Deploy 600 agent replicas:
```bash
kubectl scale statefulset opc-collector --replicas=600 -n opc-system
```

3. Distribute devices across agents:
   - Option 1: Use database device source with shard key
   - Option 2: Generate separate machine files per agent
   - Option 3: Use etcd with prefix-based distribution

## Troubleshooting

### High memory usage
- Reduce `agent.max_concurrency`
- Reduce `batch.max_memory_mb`
- Adjust `agent.gc_percent` (lower = more aggressive GC)

### Slow collection
- Increase `agent.max_concurrency`
- Check `opc_collection_duration_seconds` metric
- Verify network latency to OPC servers

### InfluxDB write errors
- Check `opc_influxdb_write_errors_total` metric
- Verify InfluxDB token and permissions
- Check InfluxDB capacity

### Connection pool exhausted
- Increase `connection_pool.opcua.max_open`
- Reduce device count per agent
- Check for slow/unresponsive OPC servers

## Performance Tuning

For optimal performance with 2,000 devices per agent:

```yaml
agent:
  max_concurrency: 100        # 100 parallel workers
  max_devices: 2000
  gc_percent: 75              # Memory vs CPU trade-off

connection_pool:
  opcua:
    max_open: 500             # Keep connections open
    max_idle: 500
    max_lifetime: 3600

batch:
  interval: 10                # 10-second batching
  max_size: 10000            # Max points per batch
```

## Next Steps

- Set up Grafana dashboards for visualization
- Configure alerting rules
- Implement device auto-discovery
- Set up high availability with multiple agents
- Configure backup and disaster recovery

## Support

For issues or questions:
- Check logs: `journalctl -u opc-collector -f`
- Review metrics: `http://localhost:9090/metrics`
- Enable debug logging: Set `logging.level: "debug"` in config
