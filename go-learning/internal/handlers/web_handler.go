package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebHandler Web 页面处理器
type WebHandler struct{}

// NewWebHandler 创建 Web 处理器实例
func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

// HomePage 主页
// GET /
func (h *WebHandler) HomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "TaskHub - Go 学习项目",
	})
}

// TasksPage 任务管理页面
// GET /tasks
func (h *WebHandler) TasksPage(c *gin.Context) {
	c.HTML(http.StatusOK, "tasks.html", gin.H{
		"title": "任务管理",
	})
}

// FilesPage 文件管理页面
// GET /files
func (h *WebHandler) FilesPage(c *gin.Context) {
	c.HTML(http.StatusOK, "files.html", gin.H{
		"title": "文件管理",
	})
}

// MonitorPage 实时监控页面
// GET /monitor
func (h *WebHandler) MonitorPage(c *gin.Context) {
	c.HTML(http.StatusOK, "monitor.html", gin.H{
		"title": "实时监控",
	})
}
