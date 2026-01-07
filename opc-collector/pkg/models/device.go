package models

import "time"

// OPCProtocol represents the type of OPC protocol
type OPCProtocol string

const (
	ProtocolOPCUA   OPCProtocol = "opcua"
	ProtocolOPCDA   OPCProtocol = "opcda"
	ProtocolGateway OPCProtocol = "gateway"
)

// HealthStatus represents the health status of a device
type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"
	HealthDegraded  HealthStatus = "degraded"
	HealthUnhealthy HealthStatus = "unhealthy"
	HealthDisabled  HealthStatus = "disabled"
)

// OPCDevice represents an OPC server/device to collect data from
type OPCDevice struct {
	ID               string           `json:"id" yaml:"id"`
	Name             string           `json:"name" yaml:"name"`
	IP               string           `json:"ip" yaml:"ip"`
	Port             int              `json:"port" yaml:"port"`
	Protocol         OPCProtocol      `json:"protocol" yaml:"protocol"`
	Enabled          bool             `json:"enabled" yaml:"enabled"`
	ConnectionConfig ConnectionConfig `json:"connection_config" yaml:"connection_config"`
	Metrics          []MetricDefinition `json:"metrics" yaml:"metrics"`
	Interval         int              `json:"interval" yaml:"interval"` // Collection interval in seconds
	HealthStatus     HealthStatus     `json:"health_status"`
	LastSuccess      time.Time        `json:"last_success"`
	FailureCount     int              `json:"failure_count"`
	Location         string           `json:"location" yaml:"location"`
	Tags             map[string]string `json:"tags" yaml:"tags"`
}

// ConnectionConfig holds connection-specific configuration for different OPC protocols
type ConnectionConfig struct {
	// OPC UA specific
	SecurityMode   string `json:"security_mode" yaml:"security_mode"`     // None, Sign, SignAndEncrypt
	SecurityPolicy string `json:"security_policy" yaml:"security_policy"` // None, Basic256Sha256
	Username       string `json:"username" yaml:"username"`
	Password       string `json:"password" yaml:"password"`
	Certificate    string `json:"certificate" yaml:"certificate"`

	// OPC DA specific
	ProgID string `json:"prog_id" yaml:"prog_id"`
	CLSID  string `json:"clsid" yaml:"clsid"`

	// Common
	Timeout       int `json:"timeout" yaml:"timeout"`             // Timeout in seconds
	RetryAttempts int `json:"retry_attempts" yaml:"retry_attempts"`
	KeepAlive     int `json:"keep_alive" yaml:"keep_alive"` // Keep-alive interval in seconds
}

// MetricDefinition defines a metric to collect from the OPC server
type MetricDefinition struct {
	NodeID      string  `json:"node_id" yaml:"node_id"`           // OPC UA NodeID or DA ItemID
	Name        string  `json:"name" yaml:"name"`                 // Friendly name
	DataType    string  `json:"data_type" yaml:"data_type"`       // float, int, string, bool
	Unit        string  `json:"unit" yaml:"unit"`                 // celsius, percent, etc
	ScaleFactor float64 `json:"scale_factor" yaml:"scale_factor"` // Scaling factor for value
}
