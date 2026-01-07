package storage

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"go.uber.org/zap"

	"opc-collector/pkg/config"
	"opc-collector/pkg/logger"
	"opc-collector/pkg/models"
)

// BatchWriter writes metrics to InfluxDB in batches
type BatchWriter struct {
	client        influxdb2.Client
	writeAPI      api.WriteAPIBlocking
	org           string
	bucket        string
	batchSize     int
	maxRetries    int
	retryInterval time.Duration
	timeout       time.Duration
	logger        *zap.Logger
}

// NewBatchWriter creates a new batch writer for InfluxDB
func NewBatchWriter(cfg config.InfluxDBConfig) (*BatchWriter, error) {
	// Create InfluxDB client
	client := influxdb2.NewClientWithOptions(
		cfg.URL,
		cfg.Token,
		influxdb2.DefaultOptions().
			SetBatchSize(uint(cfg.BatchSize)).
			SetMaxRetries(uint(cfg.MaxRetries)).
			SetRetryInterval(uint(cfg.RetryInterval)),
	)

	// Get blocking write API
	writeAPI := client.WriteAPIBlocking(cfg.Org, cfg.Bucket)

	return &BatchWriter{
		client:        client,
		writeAPI:      writeAPI,
		org:           cfg.Org,
		bucket:        cfg.Bucket,
		batchSize:     cfg.BatchSize,
		maxRetries:    cfg.MaxRetries,
		retryInterval: time.Duration(cfg.RetryInterval) * time.Second,
		timeout:       time.Duration(cfg.Timeout) * time.Second,
		logger:        logger.Named("influxdb"),
	}, nil
}

// Flush writes a batch of metrics to InfluxDB
func (w *BatchWriter) Flush(data []*models.MetricData) error {
	if len(data) == 0 {
		return nil
	}

	w.logger.Debug("starting flush",
		zap.Int("metrics_count", len(data)))

	startTime := time.Now()

	// Convert metrics to InfluxDB points
	points := make([]*write.Point, 0, len(data)*5) // Estimate 5 metrics per device

	for _, metric := range data {
		devicePoints := w.convertToPoints(metric)
		points = append(points, devicePoints...)
	}

	w.logger.Debug("converted to points",
		zap.Int("points_count", len(points)),
		zap.Duration("conversion_time", time.Since(startTime)))

	// Write in batches
	totalWritten := 0
	for i := 0; i < len(points); i += w.batchSize {
		end := i + w.batchSize
		if end > len(points) {
			end = len(points)
		}

		batch := points[i:end]

		if err := w.writeWithRetry(batch); err != nil {
			w.logger.Error("batch write failed",
				zap.Int("batch_start", i),
				zap.Int("batch_end", end),
				zap.Error(err))
			return fmt.Errorf("write batch: %w", err)
		}

		totalWritten += len(batch)
	}

	duration := time.Since(startTime)

	w.logger.Info("flush completed",
		zap.Int("points_written", totalWritten),
		zap.Duration("duration", duration),
		zap.Float64("points_per_second", float64(totalWritten)/duration.Seconds()))

	return nil
}

// convertToPoints converts MetricData to InfluxDB points
func (w *BatchWriter) convertToPoints(data *models.MetricData) []*write.Point {
	points := make([]*write.Point, 0, len(data.Metrics))

	for metricName, metricValue := range data.Metrics {
		// Build tags
		tags := make(map[string]string)
		tags["device_id"] = data.DeviceID
		tags["device_ip"] = data.DeviceIP
		tags["metric_name"] = metricName

		// Add custom tags
		for k, v := range data.Tags {
			tags[k] = v
		}

		// Build fields
		fields := map[string]interface{}{
			"value":   metricValue.Value,
			"quality": metricValue.Quality,
		}

		if metricValue.Unit != "" {
			fields["unit"] = metricValue.Unit
		}

		// Create point
		point := influxdb2.NewPoint(
			"opc_metrics",
			tags,
			fields,
			data.Timestamp,
		)

		points = append(points, point)
	}

	return points
}

// writeWithRetry writes a batch with retry logic
func (w *BatchWriter) writeWithRetry(points []*write.Point) error {
	var lastErr error

	for attempt := 0; attempt <= w.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(w.retryInterval)
			w.logger.Warn("retrying write",
				zap.Int("attempt", attempt),
				zap.Int("points", len(points)))
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), w.timeout)

		// Write points
		err := w.writeAPI.WritePoint(ctx, points...)
		cancel()

		if err == nil {
			return nil
		}

		lastErr = err

		w.logger.Warn("write attempt failed",
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", w.maxRetries),
			zap.Error(err))
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", w.maxRetries, lastErr)
}

// Health checks the health of the InfluxDB connection
func (w *BatchWriter) Health(ctx context.Context) error {
	health, err := w.client.Health(ctx)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if health.Status != "pass" {
		return fmt.Errorf("InfluxDB unhealthy: %s - %s", health.Status, *health.Message)
	}

	w.logger.Debug("health check passed",
		zap.String("status", string(health.Status)))

	return nil
}

// Close closes the InfluxDB client
func (w *BatchWriter) Close() {
	w.logger.Info("closing InfluxDB client")
	w.client.Close()
}
