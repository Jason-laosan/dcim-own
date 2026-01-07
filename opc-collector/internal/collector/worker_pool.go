package collector

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/opc-collector/internal/protocol"
	"github.com/yourusername/opc-collector/pkg/logger"
	"github.com/yourusername/opc-collector/pkg/models"
)

var (
	// ErrQueueFull is returned when the task queue is full
	ErrQueueFull = errors.New("task queue is full")
	// ErrPoolClosed is returned when the pool is closed
	ErrPoolClosed = errors.New("worker pool is closed")
)

// WorkerPool manages concurrent collection tasks using a semaphore pattern
type WorkerPool struct {
	workerCount  int
	queueSize    int
	taskQueue    chan *models.CollectionTask
	resultQueue  chan *models.MetricData
	errorQueue   chan *TaskError
	semaphore    chan struct{}
	protocols    map[models.OPCProtocol]protocol.Protocol
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	logger       *zap.Logger
	statsmu      sync.RWMutex
	stats        WorkerPoolStats
}

// WorkerPoolStats holds worker pool statistics
type WorkerPoolStats struct {
	TasksSubmitted  int64
	TasksCompleted  int64
	TasksFailed     int64
	TotalDuration   time.Duration
	ActiveWorkers   int
	QueuedTasks     int
}

// TaskError represents an error that occurred during task execution
type TaskError struct {
	Task  *models.CollectionTask
	Error error
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount, queueSize int, protocols map[models.OPCProtocol]protocol.Protocol) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		workerCount: workerCount,
		queueSize:   queueSize,
		taskQueue:   make(chan *models.CollectionTask, queueSize),
		resultQueue: make(chan *models.MetricData, queueSize*2),
		errorQueue:  make(chan *TaskError, queueSize),
		semaphore:   make(chan struct{}, workerCount),
		protocols:   protocols,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger.Named("worker-pool"),
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.logger.Info("starting worker pool",
		zap.Int("workers", wp.workerCount),
		zap.Int("queue_size", wp.queueSize))

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	workerLogger := wp.logger.With(zap.Int("worker_id", id))
	workerLogger.Debug("worker started")

	for {
		select {
		case task := <-wp.taskQueue:
			// Acquire semaphore
			wp.semaphore <- struct{}{}

			wp.statsmu.Lock()
			wp.stats.ActiveWorkers++
			wp.statsmu.Unlock()

			// Execute task
			startTime := time.Now()
			result, err := wp.executeTask(task)
			duration := time.Since(startTime)

			// Release semaphore
			<-wp.semaphore

			wp.statsmu.Lock()
			wp.stats.ActiveWorkers--
			wp.stats.TotalDuration += duration
			wp.statsmu.Unlock()

			if err != nil {
				wp.statsmu.Lock()
				wp.stats.TasksFailed++
				wp.statsmu.Unlock()

				wp.errorQueue <- &TaskError{
					Task:  task,
					Error: err,
				}

				task.RecordFailure()

				workerLogger.Error("task failed",
					zap.String("device_id", task.DeviceID),
					zap.Error(err),
					zap.Duration("duration", duration))
			} else {
				wp.statsmu.Lock()
				wp.stats.TasksCompleted++
				wp.statsmu.Unlock()

				if result != nil {
					wp.resultQueue <- result
				}

				task.RecordSuccess()

				workerLogger.Debug("task completed",
					zap.String("device_id", task.DeviceID),
					zap.Duration("duration", duration))
			}

		case <-wp.ctx.Done():
			workerLogger.Debug("worker stopped")
			return
		}
	}
}

// executeTask executes a single collection task
func (wp *WorkerPool) executeTask(task *models.CollectionTask) (*models.MetricData, error) {
	if task.Device == nil {
		return nil, errors.New("task device is nil")
	}

	// Get the appropriate protocol
	proto, ok := wp.protocols[task.Device.Protocol]
	if !ok {
		return nil, errors.New("unsupported protocol: " + string(task.Device.Protocol))
	}

	// Create context with timeout
	timeout := time.Duration(task.Device.ConnectionConfig.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(wp.ctx, timeout)
	defer cancel()

	// Collect metrics
	result, err := proto.Collect(ctx, task.Device)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Submit submits a task to the worker pool
func (wp *WorkerPool) Submit(task *models.CollectionTask) error {
	select {
	case wp.taskQueue <- task:
		wp.statsmu.Lock()
		wp.stats.TasksSubmitted++
		wp.stats.QueuedTasks = len(wp.taskQueue)
		wp.statsmu.Unlock()
		return nil

	case <-wp.ctx.Done():
		return ErrPoolClosed

	default:
		wp.logger.Warn("task queue full, dropping task",
			zap.String("device_id", task.DeviceID))
		return ErrQueueFull
	}
}

// Results returns the result queue channel
func (wp *WorkerPool) Results() <-chan *models.MetricData {
	return wp.resultQueue
}

// Errors returns the error queue channel
func (wp *WorkerPool) Errors() <-chan *TaskError {
	return wp.errorQueue
}

// Stats returns current worker pool statistics
func (wp *WorkerPool) Stats() WorkerPoolStats {
	wp.statsmu.RLock()
	defer wp.statsmu.RUnlock()

	stats := wp.stats
	stats.QueuedTasks = len(wp.taskQueue)
	return stats
}

// Stop gracefully stops the worker pool
func (wp *WorkerPool) Stop() {
	wp.logger.Info("stopping worker pool")

	// Cancel context to stop workers
	wp.cancel()

	// Wait for all workers to finish
	wp.wg.Wait()

	// Close channels
	close(wp.taskQueue)
	close(wp.resultQueue)
	close(wp.errorQueue)

	wp.logger.Info("worker pool stopped",
		zap.Int64("tasks_submitted", wp.stats.TasksSubmitted),
		zap.Int64("tasks_completed", wp.stats.TasksCompleted),
		zap.Int64("tasks_failed", wp.stats.TasksFailed))
}
