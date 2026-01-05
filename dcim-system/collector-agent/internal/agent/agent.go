package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dcim/collector-agent/internal/cache"
	"github.com/dcim/collector-agent/internal/collector"
	"github.com/dcim/collector-agent/internal/protocol"
	"github.com/dcim/collector-agent/internal/receiver"
	"github.com/dcim/collector-agent/internal/scheduler"
	"github.com/dcim/collector-agent/pkg/config"
	"github.com/dcim/collector-agent/pkg/logger"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
)

// Agent 采集Agent
type Agent struct {
	config     *config.Config
	collector  *collector.Collector
	scheduler  *scheduler.Scheduler
	receiver   *receiver.Receiver // 被动接收器
	cache      *cache.LocalCache
	mqttClient mqtt.Client
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// NewAgent 创建Agent实例
func NewAgent(cfg *config.Config) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化本地缓存
	localCache, err := cache.NewLocalCache(
		cfg.Cache.Path,
		cfg.Cache.MaxCacheTime,
		cfg.Cache.CleanInterval,
	)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create local cache: %w", err)
	}

	// 创建采集器
	coll := collector.NewCollector(localCache, cfg.Agent.MaxConcurrency)

	// 注册协议插件
	// SNMP协议
	snmpProtocol, err := protocol.NewSNMPProtocol(map[string]interface{}{
		"version":   "v2c",
		"community": "public",
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create SNMP protocol: %w", err)
	}
	coll.RegisterProtocol("snmp", snmpProtocol)

	// 创建调度器（主动拉取模式）
	var sched *scheduler.Scheduler
	if cfg.Agent.EnablePullMode {
		sched = scheduler.NewScheduler(coll)
	}

	// 创建MQTT客户端
	mqttClient := createMQTTClient(cfg.MQTT)

	agent := &Agent{
		config:     cfg,
		collector:  coll,
		scheduler:  sched,
		cache:      localCache,
		mqttClient: mqttClient,
		ctx:        ctx,
		cancel:     cancel,
	}

	// 创建被动接收器（被动接收模式）
	if cfg.Agent.EnablePushMode && cfg.Receiver.Enabled {
		rec, err := receiver.NewReceiver(&cfg.Receiver, agent.handleReceivedData)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create receiver: %w", err)
		}
		agent.receiver = rec
	}

	return agent, nil
}

// Start 启动Agent
func (a *Agent) Start() error {
	logger.Log.Info("starting agent",
		zap.String("agent_id", a.config.Agent.ID),
		zap.String("agent_name", a.config.Agent.Name))

	// 连接MQTT
	if token := a.mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect MQTT: %w", token.Error())
	}
	logger.Log.Info("MQTT connected")

	// 启动主动拉取模式
	if a.config.Agent.EnablePullMode && a.scheduler != nil {
		a.scheduler.Start()
		logger.Log.Info("pull mode started")
	}

	// 启动被动接收模式
	if a.config.Agent.EnablePushMode && a.receiver != nil {
		if err := a.receiver.Start(); err != nil {
			return fmt.Errorf("failed to start receiver: %w", err)
		}
		logger.Log.Info("push mode started")
	}

	// 启动心跳上报
	a.wg.Add(1)
	go a.heartbeatLoop()

	// 启动缓存数据重发
	a.wg.Add(1)
	go a.retryLoop()

	logger.Log.Info("agent started successfully")

	return nil
}

// Stop 停止Agent
func (a *Agent) Stop() {
	logger.Log.Info("stopping agent")

	// 停止调度器
	if a.scheduler != nil {
		a.scheduler.Stop()
	}

	// 停止接收器
	if a.receiver != nil {
		a.receiver.Stop()
	}

	// 断开MQTT连接
	a.mqttClient.Disconnect(250)

	// 取消上下文
	a.cancel()

	// 等待所有goroutine退出
	a.wg.Wait()

	// 关闭采集器
	a.collector.Close()

	// 关闭缓存
	a.cache.Close()

	logger.Log.Info("agent stopped")
}

// AddTask 添加采集任务
func (a *Agent) AddTask(task *protocol.CollectTask) error {
	return a.scheduler.AddTask(task)
}

// RemoveTask 移除采集任务
func (a *Agent) RemoveTask(taskID string) error {
	return a.scheduler.RemoveTask(taskID)
}

// PublishData 发布数据到MQTT
func (a *Agent) PublishData(data *protocol.DeviceData) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	token := a.mqttClient.Publish(a.config.MQTT.Topic, a.config.MQTT.QoS, false, payload)
	token.Wait()

	if token.Error() != nil {
		// 发布失败，缓存数据
		logger.Log.Warn("failed to publish data, caching",
			zap.String("device_id", data.DeviceID),
			zap.Error(token.Error()))

		if err := a.cache.Save(data); err != nil {
			logger.Log.Error("failed to cache data", zap.Error(err))
		}

		return token.Error()
	}

	logger.Log.Debug("data published",
		zap.String("device_id", data.DeviceID),
		zap.String("topic", a.config.MQTT.Topic))

	return nil
}

// heartbeatLoop 心跳上报循环
func (a *Agent) heartbeatLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Duration(a.config.Agent.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.sendHeartbeat()
		case <-a.ctx.Done():
			return
		}
	}
}

// sendHeartbeat 发送心跳
func (a *Agent) sendHeartbeat() {
	heartbeat := map[string]interface{}{
		"agent_id":    a.config.Agent.ID,
		"agent_name":  a.config.Agent.Name,
		"data_center": a.config.Agent.DataCenter,
		"room":        a.config.Agent.Room,
		"timestamp":   time.Now().Unix(),
		"status":      "running",
		"pull_mode":   a.config.Agent.EnablePullMode,
		"push_mode":   a.config.Agent.EnablePushMode,
	}

	payload, _ := json.Marshal(heartbeat)
	topic := fmt.Sprintf("%s/heartbeat", a.config.MQTT.Topic)

	a.mqttClient.Publish(topic, a.config.MQTT.QoS, false, payload)

	logger.Log.Debug("heartbeat sent")
}

// retryLoop 重试发送缓存数据循环
func (a *Agent) retryLoop() {
	defer a.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.retryCachedData()
		case <-a.ctx.Done():
			return
		}
	}
}

// retryCachedData 重试发送缓存数据
func (a *Agent) retryCachedData() {
	dataList, err := a.cache.GetAll()
	if err != nil {
		logger.Log.Error("failed to get cached data", zap.Error(err))
		return
	}

	if len(dataList) == 0 {
		return
	}

	logger.Log.Info("retrying cached data", zap.Int("count", len(dataList)))

	for _, data := range dataList {
		if err := a.PublishData(data); err == nil {
			// 发送成功，删除缓存
			a.cache.Delete(data.DeviceID, data.Timestamp)
		}
	}
}

// handleReceivedData 处理被动接收的数据
func (a *Agent) handleReceivedData(data *protocol.DeviceData) error {
	logger.Log.Info("received data from push mode",
		zap.String("device_id", data.DeviceID),
		zap.String("device_ip", data.DeviceIP))

	// 发布数据到MQTT
	return a.PublishData(data)
}

// createMQTTClient 创建MQTT客户端
func createMQTTClient(cfg config.MQTTConfig) mqtt.Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetClientID(cfg.ClientID)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(10 * time.Second)

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		logger.Log.Warn("MQTT connection lost", zap.Error(err))
	})

	opts.SetReconnectingHandler(func(client mqtt.Client, opts *mqtt.ClientOptions) {
		logger.Log.Info("MQTT reconnecting")
	})

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		logger.Log.Info("MQTT connected")
	})

	return mqtt.NewClient(opts)
}
