package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dcim/services/data-processor/internal/processor"
	"github.com/dcim/services/data-processor/pkg/config"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-redis/redis/v8"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 初始化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// 连接InfluxDB
	influxClient := influxdb2.NewClient(cfg.InfluxDB.URL, cfg.InfluxDB.Token)
	defer influxClient.Close()

	// 连接Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// 创建数据处理器
	dataProcessor := processor.NewDataProcessor(
		influxClient,
		cfg.InfluxDB.Org,
		cfg.InfluxDB.Bucket,
		redisClient,
		logger,
	)

	// 创建MQTT客户端
	mqttClient := createMQTTClient(cfg.MQTT, dataProcessor, logger)

	// 连接MQTT
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		logger.Fatal("failed to connect MQTT", zap.Error(token.Error()))
	}

	logger.Info("数据处理服务已启动",
		zap.String("mqtt_broker", cfg.MQTT.Broker),
		zap.String("mqtt_topic", cfg.MQTT.Topic))

	// 监听退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	// 优雅退出
	logger.Info("shutting down...")
	mqttClient.Disconnect(250)
	logger.Info("shutdown complete")
}

// createMQTTClient 创建MQTT客户端
func createMQTTClient(cfg config.MQTTConfig, processor *processor.DataProcessor, logger *zap.Logger) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetAutoReconnect(true)

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		logger.Info("MQTT connected, subscribing to topic", zap.String("topic", cfg.Topic))

		// 订阅数据Topic
		token := client.Subscribe(cfg.Topic, cfg.QoS, func(client mqtt.Client, msg mqtt.Message) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := processor.Process(ctx, msg.Payload()); err != nil {
				logger.Error("failed to process message", zap.Error(err))
			}
		})

		token.Wait()
		if token.Error() != nil {
			logger.Error("failed to subscribe", zap.Error(token.Error()))
		} else {
			logger.Info("subscribed successfully", zap.String("topic", cfg.Topic))
		}
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Warn("MQTT connection lost", zap.Error(err))
	})

	return mqtt.NewClient(opts)
}
