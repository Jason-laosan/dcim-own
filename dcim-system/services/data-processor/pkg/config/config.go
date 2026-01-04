package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Config 数据处理服务配置
type Config struct {
	MQTT     MQTTConfig     `yaml:"mqtt"`
	InfluxDB InfluxDBConfig `yaml:"influxdb"`
	Redis    RedisConfig    `yaml:"redis"`
}

// MQTTConfig MQTT配置
type MQTTConfig struct {
	Broker   string `yaml:"broker"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Topic    string `yaml:"topic"`
	QoS      byte   `yaml:"qos"`
	ClientID string `yaml:"client_id"`
}

// InfluxDBConfig InfluxDB配置
type InfluxDBConfig struct {
	URL    string `yaml:"url"`
	Token  string `yaml:"token"`
	Org    string `yaml:"org"`
	Bucket string `yaml:"bucket"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
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
