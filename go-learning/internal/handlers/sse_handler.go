package handlers

import (
	"encoding/json"
	"fmt"
	"go-learning/internal/models"
	"io"
	"log"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// SSEHandler SSE 处理器
type SSEHandler struct {
	taskService *models.TaskService
}

// NewSSEHandler 创建 SSE 处理器实例
func NewSSEHandler(taskService *models.TaskService) *SSEHandler {
	return &SSEHandler{
		taskService: taskService,
	}
}

// SSETaskUpdates 推送任务状态更新
// GET /api/sse/tasks
func (h *SSEHandler) SSETaskUpdates(c *gin.Context) {
	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 禁用 nginx 缓冲

	// 创建通道用于发送事件
	clientChan := make(chan string)
	defer close(clientChan)

	// 启动后台 goroutine 定期推送数据
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 查询各状态的任务数量
				pendingTasks, _ := h.taskService.GetByStatus("pending")
				processingTasks, _ := h.taskService.GetByStatus("processing")
				completedTasks, _ := h.taskService.GetByStatus("completed")

				data := map[string]interface{}{
					"timestamp":        time.Now().Format(time.RFC3339),
					"pending_count":    len(pendingTasks),
					"processing_count": len(processingTasks),
					"completed_count":  len(completedTasks),
				}

				jsonData, _ := json.Marshal(data)
				clientChan <- string(jsonData)

			case <-c.Request.Context().Done():
				log.Println("SSE 客户端断开连接")
				return
			}
		}
	}()

	// 持续发送事件到客户端
	c.Stream(func(w io.Writer) bool {
		select {
		case message, ok := <-clientChan:
			if !ok {
				return false
			}

			// 发送 SSE 事件
			// 格式: data: {json}\n\n
			c.SSEvent("message", message)
			return true

		case <-c.Request.Context().Done():
			return false
		}
	})
}

// SSESystemStats 推送系统状态信息
// GET /api/sse/system
func (h *SSEHandler) SSESystemStats(c *gin.Context) {
	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	clientChan := make(chan string)
	defer close(clientChan)

	// 启动后台 goroutine 推送系统统计信息
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 获取系统内存统计
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)

				// 查询处理中的任务数量
				processingTasks, _ := h.taskService.GetByStatus("processing")

				data := map[string]interface{}{
					"timestamp":        time.Now().Format(time.RFC3339),
					"goroutines":       runtime.NumGoroutine(),
					"memory_mb":        memStats.Alloc / 1024 / 1024,
					"memory_total_mb":  memStats.TotalAlloc / 1024 / 1024,
					"gc_count":         memStats.NumGC,
					"processing_tasks": len(processingTasks),
				}

				jsonData, _ := json.Marshal(data)
				clientChan <- string(jsonData)

			case <-c.Request.Context().Done():
				log.Println("SSE 客户端断开连接")
				return
			}
		}
	}()

	// 持续发送事件到客户端
	c.Stream(func(w io.Writer) bool {
		select {
		case message, ok := <-clientChan:
			if !ok {
				return false
			}

			c.SSEvent("message", message)
			return true

		case <-c.Request.Context().Done():
			return false
		}
	})
}

// ProcessTaskAsync 异步处理任务的 HTTP 处理器
// POST /api/tasks/:id/process
// 这个函数演示如何将任务提交到 Worker Pool
func ProcessTaskAsync(c *gin.Context, taskService *models.TaskService, workerPool interface{}) {
	taskID := c.Param("id")

	c.JSON(200, gin.H{
		"message": fmt.Sprintf("任务 %s 已提交到后台处理（Worker Pool）", taskID),
		"status":  "submitted",
	})
}
