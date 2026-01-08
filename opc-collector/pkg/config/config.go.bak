package config

import (
	"opc-collector/pkg/models"
)

// Config represents the complete configuration for the OPC collector agent
type Config struct {
	Agent          AgentConfig          `yaml:"agent"`
	ConnectionPool ConnectionPoolConfig `yaml:"connection_pool"`
	Batch          BatchConfig          `yaml:"batch"`
	Kafka          KafkaConfig          `yaml:"kafka"`
	Cache          CacheConfig          `yaml:"cache"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	Monitoring     MonitoringConfig     `yaml:"monitoring"`
	Logging        LoggingConfig        `yaml:"logging"`
	Devices        DevicesConfig        `yaml:"devices"`
}

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	ID                 string `yaml:"id"`
	Name               string `yaml:"name"`
	Region             string `yaml:"region"`
	Datacenter         string `yaml:"datacenter"`
	MaxConcurrency     int    `yaml:"max_concurrency"`
	MaxDevices         int    `yaml:"max_devices"`
	CollectionInterval int    `yaml:"collection_interval"` // seconds
	MaxMemoryMB        int    `yaml:"max_memory_mb"`
	GCPercent          int    `yaml:"gc_percent"`
	HeartbeatInterval  int    `yaml:"heartbeat_interval"` // seconds
	HeartbeatEndpoint  string `yaml:"heartbeat_endpoint"`
}

// ConnectionPoolConfig holds connection pooling configuration
type ConnectionPoolConfig struct {
	OPCUA   PoolConfig `yaml:"opcua"`
	OPCDA   PoolConfig `yaml:"opcda"`
	Gateway PoolConfig `yaml:"gateway"`
}

// PoolConfig holds configuration for a single connection pool
type PoolConfig struct {
	MaxIdle     int `yaml:"max_idle"`
	MaxOpen     int `yaml:"max_open"`
	MaxLifetime int `yaml:"max_lifetime"` // seconds
	IdleTimeout int `yaml:"idle_timeout"` // seconds
}

// BatchConfig holds batch processing configuration
type BatchConfig struct {
	Interval          int `yaml:"interval"`           // seconds
	MaxSize           int `yaml:"max_size"`           // max points per batch
	MaxMemoryMB       int `yaml:"max_memory_mb"`      // max memory for buffer
	FlushTimeout      int `yaml:"flush_timeout"`      // seconds
	ConcurrentFlushes int `yaml:"concurrent_flushes"` // parallel flush workers
}

// KafkaConfig holds Kafka connection configuration
type KafkaConfig struct {
	Brokers       []string `yaml:"brokers"`
	Topic         string   `yaml:"topic"`
	MaxRetries    int      `yaml:"max_retries"`
	RetryInterval int      `yaml:"retry_interval"` // seconds
	Timeout       int      `yaml:"timeout"`        // seconds
	Compression   string   `yaml:"compression"`    // none, gzip, snappy, lz4, zstd
}

// CacheConfig holds local cache configuration
type CacheConfig struct {
	Path       string `yaml:"path"`
	MaxSizeGB  int    `yaml:"max_size_gb"`
	TTL        int    `yaml:"ttl"`         // seconds
	GCInterval int    `yaml:"gc_interval"` // seconds
}

// CircuitBreakerConfig holds circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled             bool `yaml:"enabled"`
	FailureThreshold    int  `yaml:"failure_threshold"`
	SuccessThreshold    int  `yaml:"success_threshold"`
	Timeout             int  `yaml:"timeout"` // seconds
	HalfOpenMaxRequests int  `yaml:"half_open_max_requests"`
}

// MonitoringConfig holds monitoring configuration
type MonitoringConfig struct {
	Prometheus PrometheusConfig `yaml:"prometheus"`
	Pprof      PprofConfig      `yaml:"pprof"`
}

// PrometheusConfig holds Prometheus metrics configuration
type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// PprofConfig holds pprof profiling configuration
type PprofConfig struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`  // debug, info, warn, error
	Format     string `yaml:"format"` // json, console
	Output     string `yaml:"output"` // file path or stdout
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
}

// DevicesConfig holds device configuration source settings
type DevicesConfig struct {
	Source   string           `yaml:"source"` // file, database, etcd
	File     FileSourceConfig `yaml:"file"`
	Database DBSourceConfig   `yaml:"database"`
	Etcd     EtcdSourceConfig `yaml:"etcd"`
}

// FileSourceConfig holds file-based device configuration
type FileSourceConfig struct {
	Path  string `yaml:"path"`
	Watch bool   `yaml:"watch"` // hot reload
}

// DBSourceConfig holds database-based device configuration
type DBSourceConfig struct {
	DSN             string `yaml:"dsn"`
	Table           string `yaml:"table"`
	RefreshInterval int    `yaml:"refresh_interval"` // seconds
}

// EtcdSourceConfig holds etcd-based device configuration
type EtcdSourceConfig struct {
	Endpoints []string `yaml:"endpoints"`
	Prefix    string   `yaml:"prefix"`
}

// DeviceList represents a list of devices loaded from configuration
type DeviceList struct {
	Devices []models.OPCDevice `yaml:"devices"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			MaxConcurrency:     100,
			MaxDevices:         2000,
			CollectionInterval: 10,
			MaxMemoryMB:        32768,
			GCPercent:          75,
			HeartbeatInterval:  30,
		},
		ConnectionPool: ConnectionPoolConfig{
			OPCUA: PoolConfig{
				MaxIdle:     500,
				MaxOpen:     500,
				MaxLifetime: 3600,
				IdleTimeout: 600,
			},
			OPCDA: PoolConfig{
				MaxIdle:     1000,
				MaxOpen:     1000,
				MaxLifetime: 7200,
				IdleTimeout: 1200,
			},
			Gateway: PoolConfig{
				MaxIdle:     500,
				MaxOpen:     500,
				MaxLifetime: 3600,
				IdleTimeout: 600,
			},
		},
		Batch: BatchConfig{
			Interval:          10,
			MaxSize:           10000,
			MaxMemoryMB:       256,
			FlushTimeout:      5,
			ConcurrentFlushes: 4,
		},
		Kafka: KafkaConfig{
			Brokers:       []string{"localhost:9092"},
			Topic:         "opc-metrics",
			MaxRetries:    3,
			RetryInterval: 5,
			Timeout:       30,
			Compression:   "snappy",
		},
		Cache: CacheConfig{
			MaxSizeGB:  10,
			TTL:        86400,
			GCInterval: 3600,
		},
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:             true,
			FailureThreshold:    5,
			SuccessThreshold:    2,
			Timeout:             60,
			HalfOpenMaxRequests: 3,
		},
		Monitoring: MonitoringConfig{
			Prometheus: PrometheusConfig{
				Enabled: true,
				Port:    9090,
				Path:    "/metrics",
			},
			Pprof: PprofConfig{
				Enabled: false,
				Port:    6060,
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "json",
			Output:     "stdout",
			MaxSizeMB:  100,
			MaxBackups: 10,
			MaxAgeDays: 30,
		},
		Devices: DevicesConfig{
			Source: "file",
			File: FileSourceConfig{
				Watch: true,
			},
		},
	}
}
