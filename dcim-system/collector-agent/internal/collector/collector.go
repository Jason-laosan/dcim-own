package collector

import (
	"context"
	"fmt"
	"sync"

	"github.com/dcim/collector-agent/internal/cache"
	"github.com/dcim/collector-agent/internal/protocol"
	"github.com/dcim/collector-agent/pkg/logger"
	"go.uber.org/zap"
)

// Collector 数据采集器
type Collector struct {
	protocols      map[string]protocol.Protocol // 协议实例池
	cache          *cache.LocalCache            // 本地缓存
	maxConcurrency int                          // 最大并发数
	mu             sync.RWMutex                 // 读写锁
}

// NewCollector 创建采集器实例
func NewCollector(localCache *cache.LocalCache, maxConcurrency int) *Collector {
	return &Collector{
		protocols:      make(map[string]protocol.Protocol),
		cache:          localCache,
		maxConcurrency: maxConcurrency,
	}
}

// RegisterProtocol 注册协议实例
func (c *Collector) RegisterProtocol(name string, p protocol.Protocol) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.protocols[name] = p
	logger.Log.Info("protocol registered", zap.String("protocol", name))
}

// Collect 执行采集任务
func (c *Collector) Collect(ctx context.Context, task *protocol.CollectTask) (*protocol.DeviceData, error) {
	// 获取协议实例
	c.mu.RLock()
	p, exists := c.protocols[task.Protocol]
	c.mu.RUnlock()

	if !exists {
		err := fmt.Errorf("protocol not found: %s", task.Protocol)
		logger.Log.Error("protocol not found",
			zap.String("protocol", task.Protocol),
			zap.String("device_id", task.DeviceID))
		return nil, err
	}

	// 执行采集
	data, err := p.Collect(ctx, task)
	if err != nil {
		logger.Log.Error("collect failed",
			zap.String("device_id", task.DeviceID),
			zap.String("device_ip", task.DeviceIP),
			zap.Error(err))
		return data, err
	}

	logger.Log.Info("collect success",
		zap.String("device_id", task.DeviceID),
		zap.String("device_ip", task.DeviceIP),
		zap.Int("metrics_count", len(data.Metrics)))

	return data, nil
}

// CollectBatch 批量采集任务
func (c *Collector) CollectBatch(ctx context.Context, tasks []*protocol.CollectTask) []*protocol.DeviceData {
	// 使用协程池控制并发
	semaphore := make(chan struct{}, c.maxConcurrency)
	results := make([]*protocol.DeviceData, 0, len(tasks))
	resultChan := make(chan *protocol.DeviceData, len(tasks))

	var wg sync.WaitGroup

	// 启动采集任务
	for _, task := range tasks {
		wg.Add(1)
		go func(t *protocol.CollectTask) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 执行采集
			data, err := c.Collect(ctx, t)
			if err != nil {
				// 采集失败，也返回数据（包含错误信息）
				if data != nil {
					resultChan <- data
				}
				return
			}

			resultChan <- data
		}(task)
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	for data := range resultChan {
		results = append(results, data)
	}

	logger.Log.Info("batch collect completed",
		zap.Int("total_tasks", len(tasks)),
		zap.Int("success_count", len(results)))

	return results
}

// Close 关闭采集器
func (c *Collector) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 关闭所有协议实例
	for name, p := range c.protocols {
		if err := p.Close(); err != nil {
			logger.Log.Error("failed to close protocol",
				zap.String("protocol", name),
				zap.Error(err))
		}
	}

	return nil
}
