package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dcim/services/collector-mgmt/internal/handler"
	"github.com/go-redis/redis/v8"
)

// TaskService 任务服务
type TaskService struct {
	redis *redis.Client
}

// NewTaskService 创建任务服务
func NewTaskService(redisClient *redis.Client) *TaskService {
	return &TaskService{
		redis: redisClient,
	}
}

// Task 任务信息
type Task struct {
	TaskID     string                 `json:"task_id"`
	DeviceID   string                 `json:"device_id"`
	DeviceIP   string                 `json:"device_ip"`
	DeviceType string                 `json:"device_type"`
	Protocol   string                 `json:"protocol"`
	Interval   int                    `json:"interval"`
	Metrics    []string               `json:"metrics"`
	Config     map[string]interface{} `json:"config"`
	CronExpr   string                 `json:"cron_expr"`
	CreatedAt  int64                  `json:"created_at"`
}

// AgentStatus Agent状态
type AgentStatus struct {
	AgentID       string `json:"agent_id"`
	AgentName     string `json:"agent_name"`
	Status        string `json:"status"`
	LastHeartbeat int64  `json:"last_heartbeat"`
	TaskCount     int    `json:"task_count"`
	DataCenter    string `json:"data_center"`
	Room          string `json:"room"`
}

// AddTask 添加任务
func (s *TaskService) AddTask(ctx context.Context, req *handler.AddTaskRequest) error {
	task := &Task{
		TaskID:     req.TaskID,
		DeviceID:   req.DeviceID,
		DeviceIP:   req.DeviceIP,
		DeviceType: req.DeviceType,
		Protocol:   req.Protocol,
		Interval:   req.Interval,
		Metrics:    req.Metrics,
		Config:     req.Config,
		CronExpr:   req.CronExpr,
		CreatedAt:  time.Now().Unix(),
	}

	// 保存到Redis
	key := fmt.Sprintf("task:%s:%s", req.AgentID, req.TaskID)
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, key, data, 0).Err()
}

// RemoveTask 删除任务
func (s *TaskService) RemoveTask(ctx context.Context, agentID, taskID string) error {
	key := fmt.Sprintf("task:%s:%s", agentID, taskID)
	return s.redis.Del(ctx, key).Err()
}

// ListTasks 查询任务列表
func (s *TaskService) ListTasks(ctx context.Context, agentID string) ([]*Task, error) {
	pattern := fmt.Sprintf("task:%s:*", agentID)
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var task Task
		if err := json.Unmarshal([]byte(data), &task); err != nil {
			continue
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// GetAgentStatus 查询Agent状态
func (s *TaskService) GetAgentStatus(ctx context.Context, agentID string) (*AgentStatus, error) {
	key := fmt.Sprintf("agent:status:%s", agentID)
	data, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("agent not found: %s", agentID)
		}
		return nil, err
	}

	var status AgentStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return nil, err
	}

	return &status, nil
}
