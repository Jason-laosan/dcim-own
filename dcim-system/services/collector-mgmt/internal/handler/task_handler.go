package handler

import (
	"net/http"

	"github.com/dcim/services/collector-mgmt/internal/service"
	"github.com/gin-gonic/gin"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// AddTaskRequest 添加任务请求
type AddTaskRequest struct {
	AgentID    string                 `json:"agent_id" binding:"required"`
	TaskID     string                 `json:"task_id" binding:"required"`
	DeviceID   string                 `json:"device_id" binding:"required"`
	DeviceIP   string                 `json:"device_ip" binding:"required"`
	DeviceType string                 `json:"device_type" binding:"required"`
	Protocol   string                 `json:"protocol" binding:"required"`
	Interval   int                    `json:"interval" binding:"required"`
	Metrics    []string               `json:"metrics" binding:"required"`
	Config     map[string]interface{} `json:"config"`
	CronExpr   string                 `json:"cron_expr"`
}

// AddTask 添加采集任务
func (h *TaskHandler) AddTask(c *gin.Context) {
	var req AddTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.taskService.AddTask(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务添加成功",
	})
}

// RemoveTask 删除采集任务
func (h *TaskHandler) RemoveTask(c *gin.Context) {
	agentID := c.Query("agent_id")
	taskID := c.Query("task_id")

	if agentID == "" || taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id和task_id不能为空"})
		return
	}

	err := h.taskService.RemoveTask(c.Request.Context(), agentID, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "任务删除成功",
	})
}

// ListTasks 查询任务列表
func (h *TaskHandler) ListTasks(c *gin.Context) {
	agentID := c.Query("agent_id")

	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id不能为空"})
		return
	}

	tasks, err := h.taskService.ListTasks(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tasks,
	})
}

// GetAgentStatus 查询Agent状态
func (h *TaskHandler) GetAgentStatus(c *gin.Context) {
	agentID := c.Query("agent_id")

	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id不能为空"})
		return
	}

	status, err := h.taskService.GetAgentStatus(c.Request.Context(), agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// RegisterRoutes 注册路由
func (h *TaskHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.POST("/tasks", h.AddTask)
		api.DELETE("/tasks", h.RemoveTask)
		api.GET("/tasks", h.ListTasks)
		api.GET("/agent/status", h.GetAgentStatus)
	}
}
