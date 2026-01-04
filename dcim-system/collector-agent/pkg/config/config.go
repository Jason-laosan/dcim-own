package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Config 采集Agent配置
type Config struct {
	Agent AgentConfig `yaml:"agent"`
	MQTT  MQTTConfig  `yaml:"mqtt"`
	GRPC  GRPCConfig  `yaml:"grpc"`
	Cache CacheConfig `yaml:"cache"`
}

// AgentConfig Agent基础配置
type AgentConfig struct {
	ID              string `yaml:"id"`               // Agent唯一ID
	Name            string `yaml:"name"`             // Agent名称
	DataCenter      string `yaml:"data_center"`      // 所属数据中心
	Room            string `yaml:"room"`             // 所属机房
	MaxConcurrency  int    `yaml:"max_concurrency"`  // 最大并发采集数
	HeartbeatInterval int  `yaml:"heartbeat_interval"` // 心跳间隔(秒)
}

// MQTTConfig MQTT配置
type MQTTConfig struct {
	Broker   string `yaml:"broker"`    // MQTT Broker地址
	Username string `yaml:"username"`  // 用户名
	Password string `yaml:"password"`  // 密码
	Topic    string `yaml:"topic"`     // 数据上报Topic
	QoS      byte   `yaml:"qos"`       // QoS级别
	ClientID string `yaml:"client_id"` // 客户端ID
}

// GRPCConfig gRPC配置
type GRPCConfig struct {
	ServerAddr string `yaml:"server_addr"` // 服务端地址
	UseTLS     bool   `yaml:"use_tls"`     // 是否启用TLS
	CertFile   string `yaml:"cert_file"`   // 证书文件
}

// CacheConfig 本地缓存配置
type CacheConfig struct {
	Path           string `yaml:"path"`            // 缓存文件路径
	MaxCacheTime   int    `yaml:"max_cache_time"`  // 最大缓存时长(小时)
	CleanInterval  int    `yaml:"clean_interval"`  // 清理间隔(分钟)
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
