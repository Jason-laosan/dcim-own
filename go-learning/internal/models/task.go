package models

import (
	"time"

	"gorm.io/gorm"
)

// Task 任务模型结构体
type Task struct {
	ID          int        `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description" gorm:"type:text"`
	Status      string     `json:"status" gorm:"default:'pending';index"` // pending, processing, completed, failed
	Priority    int        `json:"priority" gorm:"default:1"`             // 1-5
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"type:datetime"` // 使用指针表示可选字段
}

// TaskService 任务服务，封装数据库操作
type TaskService struct {
	DB *gorm.DB
}

// NewTaskService 创建任务服务实例
func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{DB: db}
}

// GetAll 获取所有任务
func (s *TaskService) GetAll() ([]Task, error) {
	var tasks []Task
	err := s.DB.Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// GetByID 根据 ID 获取单个任务
func (s *TaskService) GetByID(id int) (*Task, error) {
	var task Task
	err := s.DB.First(&task, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 任务不存在
	}
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Create 创建新任务
func (s *TaskService) Create(task *Task) error {
	return s.DB.Create(task).Error
}

// Update 更新任务
func (s *TaskService) Update(task *Task) error {
	return s.DB.Save(task).Error
}

// UpdateStatus 更新任务状态
func (s *TaskService) UpdateStatus(id int, status string) error {
	updates := map[string]interface{}{"status": status}
	if status == "completed" {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return s.DB.Model(&Task{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除任务
func (s *TaskService) Delete(id int) error {
	return s.DB.Delete(&Task{}, id).Error
}

// GetByStatus 根据状态获取任务列表
func (s *TaskService) GetByStatus(status string) ([]Task, error) {
	var tasks []Task
	err := s.DB.Where("status = ?", status).Order("priority DESC, created_at DESC").Find(&tasks).Error
	return tasks, err
}
