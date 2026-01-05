package protocol

import (
	"context"
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
)

// SNMPProtocol SNMP协议实现
type SNMPProtocol struct {
	client *gosnmp.GoSNMP
}

// SNMPConfig SNMP配置
type SNMPConfig struct {
	Version   string `json:"version"`   // SNMP版本: v1/v2c/v3
	Community string `json:"community"` // Community字符串(v1/v2c)
	Port      uint16 `json:"port"`      // SNMP端口
	Timeout   int    `json:"timeout"`   // 超时时间(秒)
	Retries   int    `json:"retries"`   // 重试次数
}

// NewSNMPProtocol 创建SNMP协议实例
func NewSNMPProtocol(config map[string]interface{}) (*SNMPProtocol, error) {
	cfg := parseSNMPConfig(config)

	version := gosnmp.Version2c
	switch cfg.Version {
	case "v1":
		version = gosnmp.Version1
	case "v2c":
		version = gosnmp.Version2c
	case "v3":
		version = gosnmp.Version3
	}

	if cfg.Port == 0 {
		cfg.Port = 161
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5
	}
	if cfg.Retries == 0 {
		cfg.Retries = 3
	}

	client := &gosnmp.GoSNMP{
		Version:   version,
		Community: cfg.Community,
		Port:      cfg.Port,
		Timeout:   time.Duration(cfg.Timeout) * time.Second,
		Retries:   cfg.Retries,
	}

	return &SNMPProtocol{client: client}, nil
}

// Name 返回协议名称
func (s *SNMPProtocol) Name() string {
	return "SNMP"
}

// Collect 执行数据采集
func (s *SNMPProtocol) Collect(ctx context.Context, task *CollectTask) (*DeviceData, error) {
	// 设置目标设备
	s.client.Target = task.DeviceIP

	// 连接设备
	err := s.client.Connect()
	if err != nil {
		return &DeviceData{
			DeviceID:   task.DeviceID,
			DeviceIP:   task.DeviceIP,
			DeviceType: task.DeviceType,
			Timestamp:  time.Now(),
			Status:     "failed",
			Error:      fmt.Sprintf("connect failed: %v", err),
		}, err
	}
	defer s.client.Conn.Close()

	// 采集指标
	metrics := make(map[string]interface{})

	for _, metricOID := range task.Metrics {
		result, err := s.client.Get([]string{metricOID})
		if err != nil {
			metrics[metricOID] = fmt.Sprintf("error: %v", err)
			continue
		}

		for _, variable := range result.Variables {
			metrics[metricOID] = parseValue(variable)
		}
	}

	return &DeviceData{
		DeviceID:   task.DeviceID,
		DeviceIP:   task.DeviceIP,
		DeviceType: task.DeviceType,
		Timestamp:  time.Now(),
		Metrics:    metrics,
		Status:     "success",
	}, nil
}

// Validate 验证配置参数
func (s *SNMPProtocol) Validate(config map[string]interface{}) error {
	cfg := parseSNMPConfig(config)

	if cfg.Version != "v1" && cfg.Version != "v2c" && cfg.Version != "v3" {
		return fmt.Errorf("invalid SNMP version: %s", cfg.Version)
	}

	if cfg.Community == "" && cfg.Version != "v3" {
		return fmt.Errorf("community is required for SNMP %s", cfg.Version)
	}

	return nil
}

// SupportedModes 返回支持的采集模式
func (s *SNMPProtocol) SupportedModes() []CollectMode {
	// SNMP仅支持主动拉取模式
	return []CollectMode{CollectModePull}
}

// Close 关闭连接
func (s *SNMPProtocol) Close() error {
	if s.client != nil && s.client.Conn != nil {
		return s.client.Conn.Close()
	}
	return nil
}

// parseSNMPConfig 解析SNMP配置
func parseSNMPConfig(config map[string]interface{}) *SNMPConfig {
	cfg := &SNMPConfig{}

	if v, ok := config["version"].(string); ok {
		cfg.Version = v
	}
	if v, ok := config["community"].(string); ok {
		cfg.Community = v
	}
	if v, ok := config["port"].(float64); ok {
		cfg.Port = uint16(v)
	}
	if v, ok := config["timeout"].(float64); ok {
		cfg.Timeout = int(v)
	}
	if v, ok := config["retries"].(float64); ok {
		cfg.Retries = int(v)
	}

	return cfg
}

// parseValue 解析SNMP返回值
func parseValue(variable gosnmp.SnmpPDU) interface{} {
	switch variable.Type {
	case gosnmp.OctetString:
		return string(variable.Value.([]byte))
	case gosnmp.Integer:
		return variable.Value
	case gosnmp.Counter32, gosnmp.Counter64, gosnmp.Gauge32:
		return variable.Value
	default:
		return fmt.Sprintf("%v", variable.Value)
	}
}
