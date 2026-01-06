package handlers

import (
	"go-learning/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *models.TaskService
}

// NewTaskHandler 创建任务处理器实例
func NewTaskHandler(taskService *models.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// GetTasks 获取所有任务
// GET /api/tasks
func (h *TaskHandler) GetTasks(c *gin.Context) {
	// 检查是否有状态过滤参数
	status := c.Query("status")

	var tasks []models.Task
	var err error

	if status != "" {
		tasks, err = h.taskService.GetByStatus(status)
	} else {
		tasks, err = h.taskService.GetAll()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tasks",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetTask 获取单个任务
// GET /api/tasks/:id
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	task, err := h.taskService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch task",
			"details": err.Error(),
		})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CreateTask 创建新任务
// POST /api/tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task models.Task

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// 设置默认值
	if task.Status == "" {
		task.Status = "pending"
	}
	if task.Priority == 0 {
		task.Priority = 1
	}

	// 创建任务
	if err := h.taskService.Create(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask 更新任务
// PUT /api/tasks/:id
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	task.ID = id

	// 更新任务
	if err := h.taskService.Update(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask 删除任务
// DELETE /api/tasks/:id
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid task ID",
		})
		return
	}

	if err := h.taskService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}
