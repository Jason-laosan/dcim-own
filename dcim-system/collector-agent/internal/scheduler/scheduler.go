package scheduler

import (
	"context"
	"fmt"
	"sync"

	"github.com/dcim/collector-agent/internal/collector"
	"github.com/dcim/collector-agent/internal/protocol"
	"github.com/dcim/collector-agent/pkg/logger"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Scheduler 任务调度器
type Scheduler struct {
	cron      *cron.Cron
	collector *collector.Collector
	tasks     map[string]*ScheduledTask // 任务ID -> 调度任务
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

// ScheduledTask 调度任务
type ScheduledTask struct {
	Task    *protocol.CollectTask
	EntryID cron.EntryID
	Cron    string
}

// NewScheduler 创建调度器实例
func NewScheduler(c *collector.Collector) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		cron:      cron.New(cron.WithSeconds()), // 支持秒级调度
		collector: c,
		tasks:     make(map[string]*ScheduledTask),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.cron.Start()
	logger.Log.Info("scheduler started")
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.cancel()
	s.cron.Stop()
	logger.Log.Info("scheduler stopped")
}

// AddTask 添加采集任务
func (s *Scheduler) AddTask(task *protocol.CollectTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查任务是否已存在
	if _, exists := s.tasks[task.TaskID]; exists {
		return fmt.Errorf("task already exists: %s", task.TaskID)
	}

	// 生成cron表达式
	cronExpr := task.CronExpr
	if cronExpr == "" {
		// 如果没有cron表达式，使用采集间隔生成
		cronExpr = fmt.Sprintf("*/%d * * * * *", task.Interval)
	}

	// 添加定时任务
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeTask(task)
	})

	if err != nil {
		return fmt.Errorf("failed to add cron task: %w", err)
	}

	// 保存任务信息
	s.tasks[task.TaskID] = &ScheduledTask{
		Task:    task,
		EntryID: entryID,
		Cron:    cronExpr,
	}

	logger.Log.Info("task added",
		zap.String("task_id", task.TaskID),
		zap.String("device_id", task.DeviceID),
		zap.String("cron", cronExpr))

	return nil
}

// RemoveTask 移除采集任务
func (s *Scheduler) RemoveTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduledTask, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// 从cron中移除
	s.cron.Remove(scheduledTask.EntryID)

	// 从任务列表中删除
	delete(s.tasks, taskID)

	logger.Log.Info("task removed", zap.String("task_id", taskID))

	return nil
}

// UpdateTask 更新采集任务
func (s *Scheduler) UpdateTask(task *protocol.CollectTask) error {
	// 先移除旧任务
	if err := s.RemoveTask(task.TaskID); err != nil {
		return err
	}

	// 添加新任务
	return s.AddTask(task)
}

// GetTask 获取任务信息
func (s *Scheduler) GetTask(taskID string) (*ScheduledTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks 列出所有任务
func (s *Scheduler) ListTasks() []*ScheduledTask {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// executeTask 执行采集任务
func (s *Scheduler) executeTask(task *protocol.CollectTask) {
	logger.Log.Debug("executing task",
		zap.String("task_id", task.TaskID),
		zap.String("device_id", task.DeviceID))

	// 执行采集
	data, err := s.collector.Collect(s.ctx, task)
	if err != nil {
		logger.Log.Error("task execution failed",
			zap.String("task_id", task.TaskID),
			zap.Error(err))
		return
	}

	// 数据处理逻辑（上报、缓存等）由Agent层处理
	_ = data
}
