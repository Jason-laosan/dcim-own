package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Collection metrics
	CollectionTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opc_collection_total",
			Help: "Total number of collection attempts",
		},
		[]string{"device_id", "protocol", "status"},
	)

	CollectionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "opc_collection_duration_seconds",
			Help:    "Collection duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"device_id", "protocol"},
	)

	CollectionErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opc_collection_errors_total",
			Help: "Total number of collection errors",
		},
		[]string{"device_id", "protocol", "error_type"},
	)

	// Worker pool metrics
	WorkerPoolActiveWorkers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_worker_pool_active_workers",
			Help: "Number of active workers",
		},
	)

	WorkerPoolQueuedTasks = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_worker_pool_queued_tasks",
			Help: "Number of queued tasks",
		},
	)

	WorkerPoolTasksSubmitted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_worker_pool_tasks_submitted_total",
			Help: "Total number of tasks submitted",
		},
	)

	WorkerPoolTasksCompleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_worker_pool_tasks_completed_total",
			Help: "Total number of tasks completed",
		},
	)

	WorkerPoolTasksFailed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_worker_pool_tasks_failed_total",
			Help: "Total number of tasks failed",
		},
	)

	// Connection pool metrics
	ConnectionPoolActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "opc_connection_pool_active",
			Help: "Number of active connections",
		},
		[]string{"protocol"},
	)

	ConnectionPoolIdle = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "opc_connection_pool_idle",
			Help: "Number of idle connections",
		},
		[]string{"protocol"},
	)

	ConnectionPoolTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "opc_connection_pool_total",
			Help: "Total number of connections in pool",
		},
		[]string{"protocol"},
	)

	// Batch metrics
	BatchItemsReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_batch_items_received_total",
			Help: "Total number of items received by batcher",
		},
	)

	BatchItemsFlushed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_batch_items_flushed_total",
			Help: "Total number of items flushed",
		},
	)

	BatchFlushCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_batch_flush_count_total",
			Help: "Total number of batch flushes",
		},
	)

	BatchFlushErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_batch_flush_errors_total",
			Help: "Total number of batch flush errors",
		},
	)

	BatchSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "opc_batch_size",
			Help:    "Number of items in each batch",
			Buckets: []float64{100, 500, 1000, 2000, 5000, 10000, 20000, 50000},
		},
	)

	BatchFlushDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "opc_batch_flush_duration_seconds",
			Help:    "Batch flush duration in seconds",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 30, 60},
		},
	)

	BatchBufferSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_batch_buffer_size",
			Help: "Current size of batch buffer",
		},
	)

	BatchBufferMemoryMB = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_batch_buffer_memory_mb",
			Help: "Current memory usage of batch buffer in MB",
		},
	)

	// InfluxDB metrics
	InfluxDBWritesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opc_influxdb_writes_total",
			Help: "Total number of InfluxDB write attempts",
		},
		[]string{"status"},
	)

	InfluxDBPointsWritten = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_influxdb_points_written_total",
			Help: "Total number of points written to InfluxDB",
		},
	)

	InfluxDBWriteDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "opc_influxdb_write_duration_seconds",
			Help:    "InfluxDB write duration in seconds",
			Buckets: []float64{.1, .5, 1, 2, 5, 10, 30},
		},
	)

	InfluxDBWriteErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_influxdb_write_errors_total",
			Help: "Total number of InfluxDB write errors",
		},
	)

	// Cache metrics
	CacheSizeBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_cache_size_bytes",
			Help: "Size of local cache in bytes",
		},
	)

	CacheItemsStored = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_cache_items_stored_total",
			Help: "Total number of items stored in cache",
		},
	)

	// Circuit breaker metrics
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "opc_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
		},
		[]string{"device_id"},
	)

	CircuitBreakerFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opc_circuit_breaker_failures_total",
			Help: "Total number of circuit breaker failures",
		},
		[]string{"device_id"},
	)

	CircuitBreakerSuccesses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "opc_circuit_breaker_successes_total",
			Help: "Total number of circuit breaker successes",
		},
		[]string{"device_id"},
	)

	// Scheduler metrics
	SchedulerTasksCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_scheduler_tasks_count",
			Help: "Number of scheduled tasks",
		},
	)

	SchedulerTasksSubmitted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "opc_scheduler_tasks_submitted_total",
			Help: "Total number of tasks submitted by scheduler",
		},
	)

	// Agent metrics
	AgentUptime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_agent_uptime_seconds",
			Help: "Agent uptime in seconds",
		},
	)

	AgentDevicesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_agent_devices_total",
			Help: "Total number of devices managed by agent",
		},
	)

	AgentDevicesHealthy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_agent_devices_healthy",
			Help: "Number of healthy devices",
		},
	)

	AgentDevicesUnhealthy = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "opc_agent_devices_unhealthy",
			Help: "Number of unhealthy devices",
		},
	)
)
