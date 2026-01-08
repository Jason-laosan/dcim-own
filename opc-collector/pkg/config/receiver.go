package config

// ReceiverConfig holds receiver configuration for push mode
type ReceiverConfig struct {
	Enabled bool                  `yaml:"enabled"`
	HTTP    HTTPReceiverConfig    `yaml:"http"`
	MQTT    MQTTReceiverConfig    `yaml:"mqtt"`
}

// HTTPReceiverConfig holds HTTP receiver configuration
type HTTPReceiverConfig struct {
	Enabled      bool             `yaml:"enabled"`
	ListenAddr   string           `yaml:"listen_addr"`   // e.g., "0.0.0.0:8080"
	Endpoint     string           `yaml:"endpoint"`      // e.g., "/api/v1/metrics"
	ReadTimeout  int              `yaml:"read_timeout"`  // seconds
	WriteTimeout int              `yaml:"write_timeout"` // seconds
	Auth         AuthConfig       `yaml:"auth"`
}

// MQTTReceiverConfig holds MQTT receiver configuration
type MQTTReceiverConfig struct {
	Enabled          bool     `yaml:"enabled"`
	Broker           string   `yaml:"broker"`            // e.g., "tcp://localhost:1883"
	ClientID         string   `yaml:"client_id"`
	Username         string   `yaml:"username"`
	Password         string   `yaml:"password"`
	SubscribeTopics  []string `yaml:"subscribe_topics"`  // Topics to subscribe
	QoS              int      `yaml:"qos"`               // 0, 1, or 2
	KeepAlive        int      `yaml:"keep_alive"`        // seconds
	CleanSession     bool     `yaml:"clean_session"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Type     string `yaml:"type"`     // "bearer" or "basic"
	Token    string `yaml:"token"`    // for bearer auth
	Username string `yaml:"username"` // for basic auth
	Password string `yaml:"password"` // for basic auth
}

// DefaultReceiverConfig returns default receiver configuration
func DefaultReceiverConfig() ReceiverConfig {
	return ReceiverConfig{
		Enabled: false,
		HTTP: HTTPReceiverConfig{
			Enabled:      false,
			ListenAddr:   "0.0.0.0:8080",
			Endpoint:     "/api/v1/metrics",
			ReadTimeout:  30,
			WriteTimeout: 30,
			Auth: AuthConfig{
				Enabled: false,
				Type:    "bearer",
			},
		},
		MQTT: MQTTReceiverConfig{
			Enabled:         false,
			Broker:          "tcp://localhost:1883",
			ClientID:        "",
			SubscribeTopics: []string{"opc/push/+"},
			QoS:             1,
			KeepAlive:       60,
			CleanSession:    true,
		},
	}
}
