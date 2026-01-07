package batch

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/opc-collector/pkg/config"
	"github.com/yourusername/opc-collector/pkg/logger"
	"github.com/yourusername/opc-collector/pkg/models"
)

// Flusher is the interface for flushing batched data
type Flusher interface {
	Flush(data []*models.MetricData) error
}

// Batcher aggregates metrics for batch processing
type Batcher struct {
	interval      time.Duration
	maxSize       int
	maxMemoryMB   int
	buffer        *MetricBuffer
	flusher       Flusher
	inputChan     chan *models.MetricData
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	logger        *zap.Logger
	statsmu       sync.RWMutex
	stats         BatcherStats
}

// BatcherStats holds batcher statistics
type BatcherStats struct {
	ItemsReceived   int64
	ItemsFlushed    int64
	FlushCount      int64
	FlushErrors     int64
	BufferSize      int
	BufferMemoryMB  int64
}

// MetricBuffer holds metrics in memory
type MetricBuffer struct {
	mu     sync.RWMutex
	data   []*models.MetricData
	size   int
	memory int64 // estimated memory in bytes
}

// NewMetricBuffer creates a new metric buffer
func NewMetricBuffer() *MetricBuffer {
	return &MetricBuffer{
		data: make([]*models.MetricData, 0, 10000),
	}
}

// Add adds a metric to the buffer
func (mb *MetricBuffer) Add(metric *models.MetricData) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	mb.data = append(mb.data, metric)
	mb.size++

	// Rough estimate of memory usage (100 bytes per metric on average)
	mb.memory += 100
}

// Drain drains all metrics from the buffer
func (mb *MetricBuffer) Drain() []*models.MetricData {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	data := mb.data
	mb.data = make([]*models.MetricData, 0, 10000)
	mb.size = 0
	mb.memory = 0

	return data
}

// ShouldFlush returns true if the buffer should be flushed
func (mb *MetricBuffer) ShouldFlush(maxSize int, maxMemoryMB int) bool {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	if mb.size >= maxSize {
		return true
	}

	if maxMemoryMB > 0 && mb.memory > int64(maxMemoryMB)*1024*1024 {
		return true
	}

	return false
}

// Stats returns buffer statistics
func (mb *MetricBuffer) Stats() (int, int64) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.size, mb.memory
}

// NewBatcher creates a new batcher
func NewBatcher(cfg config.BatchConfig, flusher Flusher) *Batcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &Batcher{
		interval:    time.Duration(cfg.Interval) * time.Second,
		maxSize:     cfg.MaxSize,
		maxMemoryMB: cfg.MaxMemoryMB,
		buffer:      NewMetricBuffer(),
		flusher:     flusher,
		inputChan:   make(chan *models.MetricData, 10000),
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger.Named("batcher"),
	}
}

// Start starts the batcher
func (b *Batcher) Start() {
	b.logger.Info("starting batcher",
		zap.Duration("interval", b.interval),
		zap.Int("max_size", b.maxSize),
		zap.Int("max_memory_mb", b.maxMemoryMB))

	b.wg.Add(2)
	go b.collectLoop()
	go b.flushLoop()
}

// collectLoop processes incoming metrics
func (b *Batcher) collectLoop() {
	defer b.wg.Done()

	for {
		select {
		case data := <-b.inputChan:
			if data == nil {
				continue
			}

			b.buffer.Add(data)

			b.statsmu.Lock()
			b.stats.ItemsReceived++
			b.statsmu.Unlock()

			// Check if buffer should be flushed due to size/memory
			if b.buffer.ShouldFlush(b.maxSize, b.maxMemoryMB) {
				b.logger.Debug("buffer size threshold reached, flushing")
				b.flush()
			}

		case <-b.ctx.Done():
			// Final flush before shutdown
			b.logger.Info("collect loop stopping, performing final flush")
			b.flush()
			return
		}
	}
}

// flushLoop periodically flushes the buffer
func (b *Batcher) flushLoop() {
	defer b.wg.Done()

	ticker := time.NewTicker(b.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			b.flush()

		case <-b.ctx.Done():
			return
		}
	}
}

// flush flushes the buffer to the flusher
func (b *Batcher) flush() {
	data := b.buffer.Drain()
	if len(data) == 0 {
		return
	}

	b.logger.Debug("flushing batch",
		zap.Int("size", len(data)))

	startTime := time.Now()

	if err := b.flusher.Flush(data); err != nil {
		b.logger.Error("flush failed",
			zap.Int("size", len(data)),
			zap.Error(err))

		b.statsmu.Lock()
		b.stats.FlushErrors++
		b.statsmu.Unlock()

		// TODO: Write to local cache on failure
	} else {
		duration := time.Since(startTime)

		b.statsmu.Lock()
		b.stats.ItemsFlushed += int64(len(data))
		b.stats.FlushCount++
		b.statsmu.Unlock()

		b.logger.Debug("flush successful",
			zap.Int("size", len(data)),
			zap.Duration("duration", duration))
	}
}

// Add adds a metric to the batcher
func (b *Batcher) Add(metric *models.MetricData) {
	select {
	case b.inputChan <- metric:
	case <-b.ctx.Done():
	default:
		b.logger.Warn("batcher input queue full, dropping metric",
			zap.String("device_id", metric.DeviceID))
	}
}

// Stats returns batcher statistics
func (b *Batcher) Stats() BatcherStats {
	b.statsmu.RLock()
	defer b.statsmu.RUnlock()

	stats := b.stats
	bufferSize, bufferMemory := b.buffer.Stats()
	stats.BufferSize = bufferSize
	stats.BufferMemoryMB = bufferMemory / (1024 * 1024)

	return stats
}

// Stop stops the batcher
func (b *Batcher) Stop() {
	b.logger.Info("stopping batcher")

	// Cancel context
	b.cancel()

	// Wait for goroutines to finish
	b.wg.Wait()

	// Close input channel
	close(b.inputChan)

	b.logger.Info("batcher stopped",
		zap.Int64("items_received", b.stats.ItemsReceived),
		zap.Int64("items_flushed", b.stats.ItemsFlushed),
		zap.Int64("flush_count", b.stats.FlushCount),
		zap.Int64("flush_errors", b.stats.FlushErrors))
}
