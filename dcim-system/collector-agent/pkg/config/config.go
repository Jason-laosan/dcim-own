package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 采集Agent配置
type Config struct {
	Agent    AgentConfig    `yaml:"agent"`
	MQTT     MQTTConfig     `yaml:"mqtt"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Cache    CacheConfig    `yaml:"cache"`
	Receiver ReceiverConfig `yaml:"receiver"` // 被动接收配置
}

// AgentConfig Agent基础配置
type AgentConfig struct {
	ID                string   `yaml:"id"`                 // Agent唯一ID
	Name              string   `yaml:"name"`               // Agent名称
	DataCenter        string   `yaml:"data_center"`        // 所属数据中心
	Room              string   `yaml:"room"`               // 所属机房
	MaxConcurrency    int      `yaml:"max_concurrency"`    // 最大并发采集数
	HeartbeatInterval int      `yaml:"heartbeat_interval"` // 心跳间隔(秒)
	CollectModes      []string `yaml:"collect_modes"`      // 采集模式: pull(主动拉取), push(被动接收)
	EnablePullMode    bool     `yaml:"enable_pull_mode"`   // 启用主动拉取模式
	EnablePushMode    bool     `yaml:"enable_push_mode"`   // 启用被动接收模式
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
	Path          string `yaml:"path"`           // 缓存文件路径
	MaxCacheTime  int    `yaml:"max_cache_time"` // 最大缓存时长(小时)
	CleanInterval int    `yaml:"clean_interval"` // 清理间隔(分钟)
}

// ReceiverConfig 被动接收配置
type ReceiverConfig struct {
	Enabled        bool                 `yaml:"enabled"`         // 是否启用被动接收
	MQTTReceiver   MQTTReceiverConfig   `yaml:"mqtt_receiver"`   // MQTT接收器配置
	ModbusReceiver ModbusReceiverConfig `yaml:"modbus_receiver"` // Modbus接收器配置
}

// MQTTReceiverConfig MQTT接收器配置
type MQTTReceiverConfig struct {
	Enabled         bool     `yaml:"enabled"`          // 是否启用
	Broker          string   `yaml:"broker"`           // MQTT Broker地址
	Username        string   `yaml:"username"`         // 用户名
	Password        string   `yaml:"password"`         // 密码
	SubscribeTopics []string `yaml:"subscribe_topics"` // 订阅的Topic列表
	QoS             byte     `yaml:"qos"`              // QoS级别
	ClientID        string   `yaml:"client_id"`        // 客户端ID
}

// ModbusReceiverConfig Modbus接收器配置
type ModbusReceiverConfig struct {
	Enabled    bool   `yaml:"enabled"`     // 是否启用
	Mode       string `yaml:"mode"`        // 模式: tcp/rtu
	ListenAddr string `yaml:"listen_addr"` // 监听地址 (TCP模式)
	SerialPort string `yaml:"serial_port"` // 串口设备 (RTU模式)
	BaudRate   int    `yaml:"baud_rate"`   // 波特率 (RTU模式)
	SlaveID    byte   `yaml:"slave_id"`    // 从站ID
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
