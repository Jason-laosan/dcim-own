package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

// Config 服务配置
type Config struct {
	Server ServerConfig `yaml:"server"`
	GRPC   GRPCConfig   `yaml:"grpc"`
	Redis  RedisConfig  `yaml:"redis"`
}

// ServerConfig HTTP服务配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"` // debug, release
}

// GRPCConfig gRPC服务配置
type GRPCConfig struct {
	Port    int  `yaml:"port"`
	UseTLS  bool `yaml:"use_tls"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
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
