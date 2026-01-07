package agent

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"opc-collector/internal/batch"
	"opc-collector/internal/cache"
	"opc-collector/internal/collector"
	"opc-collector/internal/monitor"
	"opc-collector/internal/protocol"
	"opc-collector/internal/scheduler"
	"opc-collector/internal/storage"
	"opc-collector/pkg/config"
	"opc-collector/pkg/logger"
	"opc-collector/pkg/models"
)

// Agent is the main orchestrator for the OPC collector
type Agent struct {
	config        *config.Config
	devices       []*models.OPCDevice
	protocols     map[models.OPCProtocol]protocol.Protocol
	workerPool    *collector.WorkerPool
	scheduler     *scheduler.Scheduler
	batcher       *batch.Batcher
	kafkaProducer *storage.KafkaProducer
	cache         *cache.Cache
	startTime     time.Time
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	logger        *zap.Logger
}

// NewAgent creates a new agent
func NewAgent(cfg *config.Config, devices []*models.OPCDevice) (*Agent, error) {
	ctx, cancel := context.WithCancel(context.Background())

	agent := &Agent{
		config:    cfg,
		devices:   devices,
		protocols: make(map[models.OPCProtocol]protocol.Protocol),
		startTime: time.Now(),
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger.Named("agent").With(zap.String("agent_id", cfg.Agent.ID)),
	}

	// Set GOGC for memory optimization
	if cfg.Agent.GCPercent > 0 {
		debug.SetGCPercent(cfg.Agent.GCPercent)
		agent.logger.Info("set GOGC", zap.Int("gc_percent", cfg.Agent.GCPercent))
	}

	// Initialize components
	if err := agent.initializeComponents(); err != nil {
		cancel()
		return nil, fmt.Errorf("initialize components: %w", err)
	}

	return agent, nil
}

// initializeComponents initializes all agent components
func (a *Agent) initializeComponents() error {
	// Initialize protocols
	if err := a.initializeProtocols(); err != nil {
		return fmt.Errorf("initialize protocols: %w", err)
	}

	// Initialize Kafka producer
	kafkaProducer, err := storage.NewKafkaProducer(a.config.Kafka)
	if err != nil {
		return fmt.Errorf("create kafka producer: %w", err)
	}
	a.kafkaProducer = kafkaProducer

	// Initialize local cache
	cache, err := cache.NewCache(a.config.Cache)
	if err != nil {
		return fmt.Errorf("create cache: %w", err)
	}
	a.cache = cache

	// Initialize batcher
	a.batcher = batch.NewBatcher(a.config.Batch, kafkaProducer)

	// Initialize worker pool
	queueSize := a.config.Agent.MaxDevices * 2
	a.workerPool = collector.NewWorkerPool(
		a.config.Agent.MaxConcurrency,
		queueSize,
		a.protocols,
	)

	// Initialize scheduler
	interval := time.Duration(a.config.Agent.CollectionInterval) * time.Second
	a.scheduler = scheduler.NewScheduler(a.workerPool, interval)

	a.logger.Info("components initialized",
		zap.Int("devices", len(a.devices)),
		zap.Int("worker_pool_size", a.config.Agent.MaxConcurrency),
		zap.Duration("collection_interval", interval))

	return nil
}

// initializeProtocols initializes protocol handlers
func (a *Agent) initializeProtocols() error {
	// Initialize OPC UA protocol
	opcua := protocol.NewOPCUAProtocol(a.config.ConnectionPool.OPCUA)
	a.protocols[models.ProtocolOPCUA] = opcua

	// TODO: Initialize OPC DA and Gateway protocols
	// opcda := protocol.NewOPCDAProtocol(a.config.ConnectionPool.OPCDA)
	// a.protocols[models.ProtocolOPCDA] = opcda
	//
	// gateway := protocol.NewGatewayProtocol(a.config.ConnectionPool.Gateway)
	// a.protocols[models.ProtocolGateway] = gateway

	a.logger.Info("protocols initialized",
		zap.Int("protocol_count", len(a.protocols)))

	return nil
}

// Start starts the agent
func (a *Agent) Start() error {
	a.logger.Info("starting agent",
		zap.String("agent_id", a.config.Agent.ID),
		zap.String("name", a.config.Agent.Name))

	// Start monitoring if enabled
	if a.config.Monitoring.Prometheus.Enabled {
		go a.startPrometheusServer()
	}

	// Start batcher
	a.batcher.Start()

	// Start worker pool
	a.workerPool.Start()

	// Load devices into scheduler
	for _, device := range a.devices {
		a.scheduler.AddDevice(device)
	}

	// Start scheduler
	a.scheduler.Start()

	// Start result processor
	a.wg.Add(1)
	go a.processResults()

	// Start error processor
	a.wg.Add(1)
	go a.processErrors()

	// Start metrics updater
	a.wg.Add(1)
	go a.updateMetrics()

	a.logger.Info("agent started successfully",
		zap.Int("scheduled_tasks", a.scheduler.GetTaskCount()))

	return nil
}

// processResults processes collection results
func (a *Agent) processResults() {
	defer a.wg.Done()

	for {
		select {
		case result := <-a.workerPool.Results():
			if result != nil {
				// Add to batcher
				a.batcher.Add(result)

				// Update metrics
				monitor.BatchItemsReceived.Inc()
			}

		case <-a.ctx.Done():
			return
		}
	}
}

// processErrors processes collection errors
func (a *Agent) processErrors() {
	defer a.wg.Done()

	for {
		select {
		case taskError := <-a.workerPool.Errors():
			if taskError != nil {
				a.logger.Error("collection error",
					zap.String("device_id", taskError.Task.DeviceID),
					zap.Error(taskError.Error))

				// Update metrics
				monitor.CollectionErrors.WithLabelValues(
					taskError.Task.DeviceID,
					string(taskError.Task.Device.Protocol),
					"collection_failed",
				).Inc()
			}

		case <-a.ctx.Done():
			return
		}
	}
}

// updateMetrics periodically updates Prometheus metrics
func (a *Agent) updateMetrics() {
	defer a.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update agent uptime
			monitor.AgentUptime.Set(time.Since(a.startTime).Seconds())

			// Update device counts
			monitor.AgentDevicesTotal.Set(float64(len(a.devices)))

			// Update worker pool metrics
			poolStats := a.workerPool.Stats()
			monitor.WorkerPoolActiveWorkers.Set(float64(poolStats.ActiveWorkers))
			monitor.WorkerPoolQueuedTasks.Set(float64(poolStats.QueuedTasks))
			monitor.WorkerPoolTasksSubmitted.Add(0) // Just to ensure it exists
			monitor.WorkerPoolTasksCompleted.Add(0)
			monitor.WorkerPoolTasksFailed.Add(0)

			// Update batcher metrics
			batchStats := a.batcher.Stats()
			monitor.BatchBufferSize.Set(float64(batchStats.BufferSize))
			monitor.BatchBufferMemoryMB.Set(float64(batchStats.BufferMemoryMB))

			// Update scheduler metrics
			monitor.SchedulerTasksCount.Set(float64(a.scheduler.GetTaskCount()))

			// Update memory stats
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			a.logger.Debug("memory stats",
				zap.Uint64("alloc_mb", m.Alloc/1024/1024),
				zap.Uint64("total_alloc_mb", m.TotalAlloc/1024/1024),
				zap.Uint64("sys_mb", m.Sys/1024/1024),
				zap.Uint32("num_gc", m.NumGC))

		case <-a.ctx.Done():
			return
		}
	}
}

// startPrometheusServer starts the Prometheus metrics HTTP server
func (a *Agent) startPrometheusServer() {
	mux := http.NewServeMux()
	mux.Handle(a.config.Monitoring.Prometheus.Path, promhttp.Handler())

	addr := fmt.Sprintf(":%d", a.config.Monitoring.Prometheus.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	a.logger.Info("starting Prometheus metrics server",
		zap.String("address", addr),
		zap.String("path", a.config.Monitoring.Prometheus.Path))

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		a.logger.Error("prometheus server error", zap.Error(err))
	}
}

// Stop gracefully stops the agent
func (a *Agent) Stop() {
	a.logger.Info("stopping agent")

	// Cancel context to signal all goroutines
	a.cancel()

	// Stop scheduler first
	a.scheduler.Stop()

	// Stop worker pool
	a.workerPool.Stop()

	// Stop batcher
	a.batcher.Stop()

	// Wait for goroutines
	a.wg.Wait()

	// Close protocols
	for name, proto := range a.protocols {
		a.logger.Info("closing protocol", zap.String("protocol", string(name)))
		if err := proto.Close(); err != nil {
			a.logger.Error("error closing protocol",
				zap.String("protocol", string(name)),
				zap.Error(err))
		}
	}

	// Close Kafka producer
	if a.kafkaProducer != nil {
		a.kafkaProducer.Close()
	}

	// Close cache
	if a.cache != nil {
		a.cache.Close()
	}

	// Sync logger
	logger.Sync()

	a.logger.Info("agent stopped")
}
