package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"go.uber.org/zap"
)

// DeviceData 设备采集数据
type DeviceData struct {
	DeviceID   string                 `json:"device_id"`
	DeviceIP   string                 `json:"device_ip"`
	DeviceType string                 `json:"device_type"`
	Timestamp  time.Time              `json:"timestamp"`
	Metrics    map[string]interface{} `json:"metrics"`
	Status     string                 `json:"status"`
	Error      string                 `json:"error"`
}

// DataProcessor 数据处理器
type DataProcessor struct {
	influxWriter api.WriteAPIBlocking
	redisClient  *redis.Client
	logger       *zap.Logger
}

// NewDataProcessor 创建数据处理器
func NewDataProcessor(
	influxClient influxdb2.Client,
	org string,
	bucket string,
	redisClient *redis.Client,
	logger *zap.Logger,
) *DataProcessor {
	return &DataProcessor{
		influxWriter: influxClient.WriteAPIBlocking(org, bucket),
		redisClient:  redisClient,
		logger:       logger,
	}
}

// Process 处理采集数据
func (p *DataProcessor) Process(ctx context.Context, data []byte) error {
	var deviceData DeviceData
	if err := json.Unmarshal(data, &deviceData); err != nil {
		p.logger.Error("failed to unmarshal data", zap.Error(err))
		return err
	}

	p.logger.Info("processing device data",
		zap.String("device_id", deviceData.DeviceID),
		zap.String("device_ip", deviceData.DeviceIP),
		zap.String("status", deviceData.Status))

	// 只处理成功采集的数据
	if deviceData.Status != "success" {
		p.logger.Warn("skip failed data",
			zap.String("device_id", deviceData.DeviceID),
			zap.String("error", deviceData.Error))
		return nil
	}

	// 写入InfluxDB
	if err := p.writeToInfluxDB(ctx, &deviceData); err != nil {
		p.logger.Error("failed to write to influxdb", zap.Error(err))
		return err
	}

	// 更新Redis缓存（用于监控大屏）
	if err := p.updateRedisCache(ctx, &deviceData); err != nil {
		p.logger.Error("failed to update redis cache", zap.Error(err))
		// Redis失败不影响主流程
	}

	return nil
}

// writeToInfluxDB 写入InfluxDB
func (p *DataProcessor) writeToInfluxDB(ctx context.Context, data *DeviceData) error {
	// 创建数据点
	point := influxdb2.NewPoint(
		"device_metrics",
		map[string]string{
			"device_id":   data.DeviceID,
			"device_ip":   data.DeviceIP,
			"device_type": data.DeviceType,
		},
		data.Metrics,
		data.Timestamp,
	)

	// 写入数据
	return p.influxWriter.WritePoint(ctx, point)
}

// updateRedisCache 更新Redis缓存
func (p *DataProcessor) updateRedisCache(ctx context.Context, data *DeviceData) error {
	key := fmt.Sprintf("device:latest:%s", data.DeviceID)

	cacheData := map[string]interface{}{
		"device_id":   data.DeviceID,
		"device_ip":   data.DeviceIP,
		"device_type": data.DeviceType,
		"timestamp":   data.Timestamp.Unix(),
		"metrics":     data.Metrics,
	}

	jsonData, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	// 缓存最新数据，TTL 5分钟
	return p.redisClient.Set(ctx, key, jsonData, 5*time.Minute).Err()
}
