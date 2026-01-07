package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"opc-collector/pkg/config"
	"opc-collector/pkg/logger"
	"opc-collector/pkg/models"
)

type KafkaProducer struct {
	producer      sarama.SyncProducer
	topic         string
	logger        *zap.Logger
	maxRetries    int
	retryInterval time.Duration
}

func NewKafkaProducer(cfg config.KafkaConfig) (*KafkaProducer, error) {
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.RequiredAcks = sarama.WaitForAll
	kafkaConfig.Producer.Retry.Max = cfg.MaxRetries
	kafkaConfig.Producer.Return.Successes = true
	kafkaConfig.Producer.Compression = sarama.CompressionSnappy
	kafkaConfig.Producer.Timeout = time.Duration(cfg.Timeout) * time.Second
	kafkaConfig.Version = sarama.V2_6_0_0

	producer, err := sarama.NewSyncProducer(cfg.Brokers, kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &KafkaProducer{
		producer:      producer,
		topic:         cfg.Topic,
		logger:        logger.Named("kafka"),
		maxRetries:    cfg.MaxRetries,
		retryInterval: time.Duration(cfg.RetryInterval) * time.Second,
	}, nil
}

func (k *KafkaProducer) Flush(data []*models.MetricData) error {
	if len(data) == 0 {
		return nil
	}

	k.logger.Debug("starting flush to kafka",
		zap.Int("metrics_count", len(data)))

	startTime := time.Now()
	successCount := 0
	errorCount := 0

	for _, metric := range data {
		if err := k.sendMetric(metric); err != nil {
			k.logger.Error("failed to send metric to kafka",
				zap.String("device_id", metric.DeviceID),
				zap.Error(err))
			errorCount++
		} else {
			successCount++
		}
	}

	duration := time.Since(startTime)

	k.logger.Info("flush completed",
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
		zap.Duration("duration", duration),
		zap.Float64("metrics_per_second", float64(successCount)/duration.Seconds()))

	if errorCount > 0 {
		return fmt.Errorf("failed to send %d out of %d metrics", errorCount, len(data))
	}

	return nil
}

func (k *KafkaProducer) sendMetric(metric *models.MetricData) error {
	messageBytes, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		Key:   sarama.StringEncoder(metric.DeviceID),
		Value: sarama.ByteEncoder(messageBytes),
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("device_id"),
				Value: []byte(metric.DeviceID),
			},
			{
				Key:   []byte("device_ip"),
				Value: []byte(metric.DeviceIP),
			},
			{
				Key:   []byte("timestamp"),
				Value: []byte(metric.Timestamp.Format(time.RFC3339)),
			},
		},
	}

	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to kafka: %w", err)
	}

	k.logger.Debug("message sent to kafka",
		zap.String("device_id", metric.DeviceID),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

func (k *KafkaProducer) Health(ctx context.Context) error {
	testMetric := &models.MetricData{
		DeviceID:  "health-check",
		DeviceIP:  "0.0.0.0",
		Timestamp: time.Now(),
		Metrics:   make(map[string]models.MetricValue),
		Tags:      map[string]string{"type": "health-check"},
		Quality:   models.QualityGood,
	}

	if err := k.sendMetric(testMetric); err != nil {
		return fmt.Errorf("kafka health check failed: %w", err)
	}

	k.logger.Debug("kafka health check passed")
	return nil
}

func (k *KafkaProducer) Close() {
	k.logger.Info("closing kafka producer")
	if err := k.producer.Close(); err != nil {
		k.logger.Error("error closing kafka producer", zap.Error(err))
	}
}
