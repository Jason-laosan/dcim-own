package scheduler

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/yourusername/opc-collector/internal/collector"
	"github.com/yourusername/opc-collector/pkg/logger"
	"github.com/yourusername/opc-collector/pkg/models"
)

// Scheduler manages the scheduling of collection tasks
type Scheduler struct {
	tasks       map[string]*models.CollectionTask
	workerPool  *collector.WorkerPool
	interval    time.Duration
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	logger      *zap.Logger
}

// NewScheduler creates a new scheduler
func NewScheduler(workerPool *collector.WorkerPool, interval time.Duration) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		tasks:      make(map[string]*models.CollectionTask),
		workerPool: workerPool,
		interval:   interval,
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger.Named("scheduler"),
	}
}

// AddTask adds a task to the scheduler
func (s *Scheduler) AddTask(task *models.CollectionTask) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tasks[task.TaskID] = task

	s.logger.Debug("task added",
		zap.String("task_id", task.TaskID),
		zap.String("device_id", task.DeviceID),
		zap.Duration("interval", task.Interval))
}

// RemoveTask removes a task from the scheduler
func (s *Scheduler) RemoveTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, taskID)

	s.logger.Debug("task removed",
		zap.String("task_id", taskID))
}

// AddDevice adds a device as a collection task
func (s *Scheduler) AddDevice(device *models.OPCDevice) {
	if !device.Enabled {
		return
	}

	interval := time.Duration(device.Interval) * time.Second
	if interval == 0 {
		interval = s.interval
	}

	task := &models.CollectionTask{
		TaskID:   "task-" + device.ID,
		DeviceID: device.ID,
		Device:   device,
		Interval: interval,
		NextRun:  time.Now(), // Start immediately
		Enabled:  true,
	}

	s.AddTask(task)
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.logger.Info("starting scheduler",
		zap.Duration("interval", s.interval),
		zap.Int("tasks", len(s.tasks)))

	s.wg.Add(1)
	go s.scheduleLoop()
}

// scheduleLoop is the main scheduling loop
func (s *Scheduler) scheduleLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAndSubmitTasks()

		case <-s.ctx.Done():
			s.logger.Info("scheduler stopping")
			return
		}
	}
}

// checkAndSubmitTasks checks which tasks should run and submits them
func (s *Scheduler) checkAndSubmitTasks() {
	s.mu.RLock()
	tasksToRun := make([]*models.CollectionTask, 0)

	for _, task := range s.tasks {
		if task.ShouldRun() {
			tasksToRun = append(tasksToRun, task)
		}
	}
	s.mu.RUnlock()

	// Submit tasks outside the lock to avoid blocking
	for _, task := range tasksToRun {
		if err := s.workerPool.Submit(task); err != nil {
			s.logger.Warn("failed to submit task",
				zap.String("task_id", task.TaskID),
				zap.String("device_id", task.DeviceID),
				zap.Error(err))
		} else {
			s.logger.Debug("task submitted",
				zap.String("task_id", task.TaskID),
				zap.String("device_id", task.DeviceID))
		}
	}
}

// GetTaskCount returns the number of tasks
func (s *Scheduler) GetTaskCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks)
}

// GetTasks returns a copy of all tasks
func (s *Scheduler) GetTasks() []*models.CollectionTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*models.CollectionTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.logger.Info("stopping scheduler")

	s.cancel()
	s.wg.Wait()

	s.logger.Info("scheduler stopped",
		zap.Int("tasks", len(s.tasks)))
}
