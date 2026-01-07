package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load loads configuration from a YAML file
func Load(path string) (*Config, error) {
	// Start with default config
	cfg := DefaultConfig()

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

// LoadDevices loads device configuration from a file
func LoadDevices(path string) (*DeviceList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read devices file: %w", err)
	}

	var devices DeviceList
	if err := yaml.Unmarshal(data, &devices); err != nil {
		return nil, fmt.Errorf("parse devices file: %w", err)
	}

	return &devices, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Agent validation
	if c.Agent.MaxConcurrency <= 0 {
		return fmt.Errorf("agent.max_concurrency must be positive")
	}
	if c.Agent.MaxDevices <= 0 {
		return fmt.Errorf("agent.max_devices must be positive")
	}
	if c.Agent.CollectionInterval <= 0 {
		return fmt.Errorf("agent.collection_interval must be positive")
	}

	// Batch validation
	if c.Batch.Interval <= 0 {
		return fmt.Errorf("batch.interval must be positive")
	}
	if c.Batch.MaxSize <= 0 {
		return fmt.Errorf("batch.max_size must be positive")
	}

	// Kafka validation
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("kafka.brokers is required")
	}
	if c.Kafka.Topic == "" {
		return fmt.Errorf("kafka.topic is required")
	}

	// Cache validation
	if c.Cache.Path == "" {
		return fmt.Errorf("cache.path is required")
	}

	// Devices validation
	if c.Devices.Source == "" {
		return fmt.Errorf("devices.source is required")
	}
	if c.Devices.Source == "file" && c.Devices.File.Path == "" {
		return fmt.Errorf("devices.file.path is required when source is file")
	}

	// Logging validation
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("logging.level must be one of: debug, info, warn, error")
	}

	validFormats := map[string]bool{"json": true, "console": true}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("logging.format must be one of: json, console")
	}

	return nil
}

// Save saves the configuration to a YAML file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
