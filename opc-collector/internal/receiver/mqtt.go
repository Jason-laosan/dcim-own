package receiver

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"opc-collector/pkg/config"
	"opc-collector/pkg/logger"
	"opc-collector/pkg/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// MQTTReceiver MQTT接收器
// 订阅MQTT主题接收下游主动推送的数据
type MQTTReceiver struct {
	config       *config.MQTTReceiverConfig
	client       mqtt.Client
	dataHandler  DataHandler
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	logger       *zap.Logger

	// 统计指标
	receivedCount uint64
	errorCount    uint64
}

// NewMQTTReceiver 创建MQTT接收器实例
func NewMQTTReceiver(cfg *config.MQTTReceiverConfig, handler DataHandler) (*MQTTReceiver, error) {
	ctx, cancel := context.WithCancel(context.Background())

	r := &MQTTReceiver{
		config:      cfg,
		dataHandler: handler,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger.Log,
	}

	return r, nil
}

// Start 启动MQTT接收器
func (r *MQTTReceiver) Start() error {
	opts := mqtt.NewClientOptions()

	// 添加Broker地址
	opts.AddBroker(r.config.Broker)

	// 设置客户端ID
	clientID := r.config.ClientID
	if clientID == "" {
		clientID = fmt.Sprintf("opc-collector-receiver-%d", time.Now().Unix())
	}
	opts.SetClientID(clientID)

	// 设置认证
	if r.config.Username != "" {
		opts.SetUsername(r.config.Username)
		opts.SetPassword(r.config.Password)
	}

	// 设置自动重连
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetMaxReconnectInterval(60 * time.Second)

	// 设置Keep Alive
	opts.SetKeepAlive(time.Duration(r.config.KeepAlive) * time.Second)

	// 设置Clean Session
	opts.SetCleanSession(r.config.CleanSession)

	// 连接丢失处理
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		r.logger.Warn("MQTT receiver connection lost",
			zap.Error(err),
			zap.String("broker", r.config.Broker))
	})

	// 连接成功处理
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		r.logger.Info("MQTT receiver connected", zap.String("broker", r.config.Broker))

		// 订阅所有配置的Topic
		for _, topic := range r.config.SubscribeTopics {
			qos := byte(r.config.QoS)
			token := client.Subscribe(topic, qos, r.messageHandler)

			if token.Wait() && token.Error() != nil {
				r.logger.Error("failed to subscribe topic",
					zap.String("topic", topic),
					zap.Error(token.Error()))
			} else {
				r.logger.Info("subscribed to topic",
					zap.String("topic", topic),
					zap.Int("qos", r.config.QoS))
			}
		}
	})

	// 重连处理
	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		r.logger.Info("MQTT receiver reconnecting...", zap.String("broker", r.config.Broker))
	})

	// 创建客户端
	r.client = mqtt.NewClient(opts)

	// 连接到Broker
	r.logger.Info("connecting to MQTT broker", zap.String("broker", r.config.Broker))
	token := r.client.Connect()

	// 等待连接完成
	if !token.WaitTimeout(30 * time.Second) {
		return fmt.Errorf("connection timeout to MQTT broker: %s", r.config.Broker)
	}

	if token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return nil
}

// Stop 停止MQTT接收器
func (r *MQTTReceiver) Stop() {
	r.logger.Info("stopping MQTT receiver")

	if r.client != nil && r.client.IsConnected() {
		// 取消订阅所有主题
		for _, topic := range r.config.SubscribeTopics {
			if token := r.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
				r.logger.Warn("failed to unsubscribe topic",
					zap.String("topic", topic),
					zap.Error(token.Error()))
			}
		}

		// 断开连接
		r.client.Disconnect(250)
	}

	r.cancel()
	r.wg.Wait()

	r.logger.Info("MQTT receiver stopped")
}

// Name 返回接收器名称
func (r *MQTTReceiver) Name() string {
	return "mqtt"
}

// messageHandler MQTT消息处理器
func (r *MQTTReceiver) messageHandler(client mqtt.Client, msg mqtt.Message) {
	r.logger.Debug("received MQTT message",
		zap.String("topic", msg.Topic()),
		zap.Int("payload_size", len(msg.Payload())),
		zap.Int("qos", int(msg.Qos())))

	// 解析消息为MetricData
	var data models.MetricData
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		r.logger.Error("failed to unmarshal MQTT message",
			zap.String("topic", msg.Topic()),
			zap.String("payload", string(msg.Payload())),
			zap.Error(err))
		atomic.AddUint64(&r.errorCount, 1)
		return
	}

	// 验证必要字段
	if data.DeviceID == "" {
		r.logger.Warn("received MQTT message without device_id",
			zap.String("topic", msg.Topic()))
		atomic.AddUint64(&r.errorCount, 1)
		return
	}

	// 如果时间戳为空，使用当前时间
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	r.logger.Debug("processing MQTT metric data",
		zap.String("device_id", data.DeviceID),
		zap.String("topic", msg.Topic()),
		zap.Int("metric_count", len(data.Metrics)))

	// 调用数据处理回调
	if err := r.dataHandler(&data); err != nil {
		r.logger.Error("failed to handle MQTT data",
			zap.String("device_id", data.DeviceID),
			zap.String("topic", msg.Topic()),
			zap.Error(err))
		atomic.AddUint64(&r.errorCount, 1)
		return
	}

	atomic.AddUint64(&r.receivedCount, 1)

	r.logger.Debug("successfully processed MQTT metric data",
		zap.String("device_id", data.DeviceID),
		zap.String("topic", msg.Topic()))
}

// GetReceivedCount 获取接收数据计数
func (r *MQTTReceiver) GetReceivedCount() uint64 {
	return atomic.LoadUint64(&r.receivedCount)
}

// GetErrorCount 获取错误计数
func (r *MQTTReceiver) GetErrorCount() uint64 {
	return atomic.LoadUint64(&r.errorCount)
}
