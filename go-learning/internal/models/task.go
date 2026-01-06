package models

import (
	"database/sql"
	"time"
)

// Task 任务模型结构体
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`       // pending, processing, completed, failed
	Priority    int       `json:"priority"`     // 1-5
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"` // 使用指针表示可选字段
}

// TaskService 任务服务，封装数据库操作
type TaskService struct {
	DB *sql.DB
}

// NewTaskService 创建任务服务实例
func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{DB: db}
}

// GetAll 获取所有任务
func (s *TaskService) GetAll() ([]Task, error) {
	query := `SELECT id, title, description, status, priority, created_at, updated_at, completed_at
	          FROM tasks ORDER BY created_at DESC`

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// GetByID 根据 ID 获取单个任务
func (s *TaskService) GetByID(id int) (*Task, error) {
	query := `SELECT id, title, description, status, priority, created_at, updated_at, completed_at
	          FROM tasks WHERE id = ?`

	var task Task
	err := s.DB.QueryRow(query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // 任务不存在
	}
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// Create 创建新任务
func (s *TaskService) Create(task *Task) error {
	query := `INSERT INTO tasks (title, description, status, priority)
	          VALUES (?, ?, ?, ?)`

	result, err := s.DB.Exec(query, task.Title, task.Description, task.Status, task.Priority)
	if err != nil {
		return err
	}

	// 获取新插入记录的 ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = int(id)

	return nil
}

// Update 更新任务
func (s *TaskService) Update(task *Task) error {
	query := `UPDATE tasks
	          SET title = ?, description = ?, status = ?, priority = ?, updated_at = CURRENT_TIMESTAMP
	          WHERE id = ?`

	_, err := s.DB.Exec(query, task.Title, task.Description, task.Status, task.Priority, task.ID)
	return err
}

// UpdateStatus 更新任务状态
func (s *TaskService) UpdateStatus(id int, status string) error {
	query := `UPDATE tasks
	          SET status = ?, updated_at = CURRENT_TIMESTAMP,
	              completed_at = CASE WHEN ? = 'completed' THEN CURRENT_TIMESTAMP ELSE completed_at END
	          WHERE id = ?`

	_, err := s.DB.Exec(query, status, status, id)
	return err
}

// Delete 删除任务
func (s *TaskService) Delete(id int) error {
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	return err
}

// GetByStatus 根据状态获取任务列表
func (s *TaskService) GetByStatus(status string) ([]Task, error) {
	query := `SELECT id, title, description, status, priority, created_at, updated_at, completed_at
	          FROM tasks WHERE status = ? ORDER BY priority DESC, created_at DESC`

	rows, err := s.DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
