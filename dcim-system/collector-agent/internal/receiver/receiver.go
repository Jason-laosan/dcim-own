package receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/dcim/collector-agent/internal/protocol"
	"github.com/dcim/collector-agent/pkg/config"
	"github.com/dcim/collector-agent/pkg/logger"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// DataHandler 数据处理回调函数
type DataHandler func(*protocol.DeviceData) error

// Receiver 被动接收器管理器
type Receiver struct {
	config         *config.ReceiverConfig
	mqttReceiver   *MQTTReceiver
	modbusReceiver *ModbusReceiver
	dataHandler    DataHandler
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// NewReceiver 创建接收器实例
func NewReceiver(cfg *config.ReceiverConfig, handler DataHandler) (*Receiver, error) {
	ctx, cancel := context.WithCancel(context.Background())

	r := &Receiver{
		config:      cfg,
		dataHandler: handler,
		ctx:         ctx,
		cancel:      cancel,
	}

	return r, nil
}

// Start 启动接收器
func (r *Receiver) Start() error {
	if !r.config.Enabled {
		logger.Log.Info("receiver is disabled")
		return nil
	}

	logger.Log.Info("starting receiver")

	// 启动MQTT接收器
	if r.config.MQTTReceiver.Enabled {
		mqttReceiver, err := NewMQTTReceiver(r.config.MQTTReceiver, r.dataHandler)
		if err != nil {
			return fmt.Errorf("failed to create MQTT receiver: %w", err)
		}
		r.mqttReceiver = mqttReceiver

		if err := r.mqttReceiver.Start(); err != nil {
			return fmt.Errorf("failed to start MQTT receiver: %w", err)
		}
		logger.Log.Info("MQTT receiver started")
	}

	// 启动Modbus接收器
	if r.config.ModbusReceiver.Enabled {
		modbusReceiver, err := NewModbusReceiver(r.config.ModbusReceiver, r.dataHandler)
		if err != nil {
			return fmt.Errorf("failed to create Modbus receiver: %w", err)
		}
		r.modbusReceiver = modbusReceiver

		if err := r.modbusReceiver.Start(); err != nil {
			return fmt.Errorf("failed to start Modbus receiver: %w", err)
		}
		logger.Log.Info("Modbus receiver started")
	}

	logger.Log.Info("receiver started successfully")
	return nil
}

// Stop 停止接收器
func (r *Receiver) Stop() {
	logger.Log.Info("stopping receiver")

	if r.mqttReceiver != nil {
		r.mqttReceiver.Stop()
	}

	if r.modbusReceiver != nil {
		r.modbusReceiver.Stop()
	}

	r.cancel()
	r.wg.Wait()

	logger.Log.Info("receiver stopped")
}

// MQTTReceiver MQTT接收器
type MQTTReceiver struct {
	config      config.MQTTReceiverConfig
	client      mqtt.Client
	dataHandler DataHandler
}

// NewMQTTReceiver 创建MQTT接收器
func NewMQTTReceiver(cfg config.MQTTReceiverConfig, handler DataHandler) (*MQTTReceiver, error) {
	return &MQTTReceiver{
		config:      cfg,
		dataHandler: handler,
	}, nil
}

// Start 启动MQTT接收器
func (m *MQTTReceiver) Start() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.config.Broker)
	opts.SetClientID(m.config.ClientID)
	opts.SetUsername(m.config.Username)
	opts.SetPassword(m.config.Password)
	opts.SetAutoReconnect(true)

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Log.Warn("MQTT receiver connection lost", zap.Error(err))
	})

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		logger.Log.Info("MQTT receiver connected")
		// 订阅所有配置的Topic
		for _, topic := range m.config.SubscribeTopics {
			if token := client.Subscribe(topic, m.config.QoS, m.messageHandler); token.Wait() && token.Error() != nil {
				logger.Log.Error("failed to subscribe topic",
					zap.String("topic", topic),
					zap.Error(token.Error()))
			} else {
				logger.Log.Info("subscribed to topic", zap.String("topic", topic))
			}
		}
	})

	m.client = mqtt.NewClient(opts)

	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect MQTT: %w", token.Error())
	}

	return nil
}

// Stop 停止MQTT接收器
func (m *MQTTReceiver) Stop() {
	if m.client != nil {
		m.client.Disconnect(250)
	}
}

// messageHandler MQTT消息处理器
func (m *MQTTReceiver) messageHandler(client mqtt.Client, msg mqtt.Message) {
	logger.Log.Debug("received MQTT message",
		zap.String("topic", msg.Topic()),
		zap.Int("payload_size", len(msg.Payload())))

	// 解析消息为DeviceData
	var data protocol.DeviceData
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		logger.Log.Error("failed to unmarshal MQTT message",
			zap.String("topic", msg.Topic()),
			zap.Error(err))
		return
	}

	// 调用数据处理回调
	if err := m.dataHandler(&data); err != nil {
		logger.Log.Error("failed to handle MQTT data",
			zap.String("device_id", data.DeviceID),
			zap.Error(err))
	}
}

// ModbusReceiver Modbus接收器
type ModbusReceiver struct {
	config      config.ModbusReceiverConfig
	dataHandler DataHandler
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewModbusReceiver 创建Modbus接收器
func NewModbusReceiver(cfg config.ModbusReceiverConfig, handler DataHandler) (*ModbusReceiver, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &ModbusReceiver{
		config:      cfg,
		dataHandler: handler,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start 启动Modbus接收器
func (m *ModbusReceiver) Start() error {
	logger.Log.Info("starting Modbus receiver",
		zap.String("mode", m.config.Mode),
		zap.String("listen_addr", m.config.ListenAddr))

	// TODO: 实现Modbus Slave模式监听
	// 这里需要根据具体的Modbus库实现
	// 1. TCP模式：监听指定端口，接收Modbus TCP请求
	// 2. RTU模式：监听串口，接收Modbus RTU请求

	logger.Log.Warn("Modbus receiver is not fully implemented yet")
	return nil
}

// Stop 停止Modbus接收器
func (m *ModbusReceiver) Stop() {
	m.cancel()
	logger.Log.Info("Modbus receiver stopped")
}
