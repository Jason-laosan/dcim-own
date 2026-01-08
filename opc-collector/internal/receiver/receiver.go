package receiver

import (
	"context"
	"fmt"
	"sync"

	"opc-collector/pkg/config"
	"opc-collector/pkg/logger"
	"opc-collector/pkg/models"

	"go.uber.org/zap"
)

// DataHandler 数据处理回调函数
// 接收从下游推送过来的数据，并进行处理
type DataHandler func(*models.MetricData) error

// Receiver 接收器接口
// 定义所有接收器必须实现的方法
type Receiver interface {
	Start() error
	Stop()
	Name() string
}

// Manager 接收器管理器
// 管理所有类型的接收器（HTTP、MQTT等）
type Manager struct {
	config      *config.ReceiverConfig
	receivers   []Receiver
	dataHandler DataHandler
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
	logger      *zap.Logger
}

// NewManager 创建接收器管理器实例
func NewManager(cfg *config.ReceiverConfig, handler DataHandler) (*Manager, error) {
	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		config:      cfg,
		dataHandler: handler,
		ctx:         ctx,
		cancel:      cancel,
		receivers:   make([]Receiver, 0),
		logger:      logger.Log,
	}

	return m, nil
}

// Start 启动所有配置的接收器
func (m *Manager) Start() error {
	if !m.config.Enabled {
		m.logger.Info("receiver manager is disabled")
		return nil
	}

	m.logger.Info("starting receiver manager")

	m.mu.Lock()
	defer m.mu.Unlock()

	// 启动HTTP接收器
	if m.config.HTTP.Enabled {
		httpReceiver, err := NewHTTPReceiver(&m.config.HTTP, m.dataHandler)
		if err != nil {
			return fmt.Errorf("failed to create HTTP receiver: %w", err)
		}

		if err := httpReceiver.Start(); err != nil {
			return fmt.Errorf("failed to start HTTP receiver: %w", err)
		}

		m.receivers = append(m.receivers, httpReceiver)
		m.logger.Info("HTTP receiver started", zap.String("address", m.config.HTTP.ListenAddr))
	}

	// 启动MQTT接收器
	if m.config.MQTT.Enabled {
		mqttReceiver, err := NewMQTTReceiver(&m.config.MQTT, m.dataHandler)
		if err != nil {
			return fmt.Errorf("failed to create MQTT receiver: %w", err)
		}

		if err := mqttReceiver.Start(); err != nil {
			return fmt.Errorf("failed to start MQTT receiver: %w", err)
		}

		m.receivers = append(m.receivers, mqttReceiver)
		m.logger.Info("MQTT receiver started",
			zap.String("broker", m.config.MQTT.Broker),
			zap.Strings("topics", m.config.MQTT.SubscribeTopics))
	}

	m.logger.Info("receiver manager started successfully", zap.Int("active_receivers", len(m.receivers)))
	return nil
}

// Stop 停止所有接收器
func (m *Manager) Stop() {
	m.logger.Info("stopping receiver manager")

	m.mu.Lock()
	defer m.mu.Unlock()

	// 停止所有接收器
	for _, receiver := range m.receivers {
		m.logger.Info("stopping receiver", zap.String("name", receiver.Name()))
		receiver.Stop()
	}

	m.cancel()
	m.wg.Wait()

	m.logger.Info("receiver manager stopped")
}

// GetActiveReceivers 获取活跃的接收器数量
func (m *Manager) GetActiveReceivers() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.receivers)
}
