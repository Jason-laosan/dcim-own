package main

import (
	"context"
	"go-learning/internal/database"
	"go-learning/internal/handlers"
	"go-learning/internal/middleware"
	"go-learning/internal/models"
	"go-learning/internal/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("=== TaskHub - Go 学习项目 ===")
	log.Println("启动中...")

	// 1. 初始化数据库
	log.Println("初始化数据库...")
	if err := database.InitDB("tasks.db"); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer database.CloseDB()

	// 2. 创建服务实例
	db := database.GetDB()
	taskService := models.NewTaskService(db)
	fileService := models.NewFileService(db)

	// 3. 创建 Worker Pool
	log.Println("启动 Worker Pool...")
	workerPool := services.NewWorkerPool(5, func(taskID int, status string) {
		// 更新任务状态的回调函数
		if err := taskService.UpdateStatus(taskID, status); err != nil {
			log.Printf("更新任务 %d 状态失败: %v", taskID, err)
		}
	})
	workerPool.Start()
	defer workerPool.Shutdown()

	// 4. 创建处理器实例
	taskHandler := handlers.NewTaskHandler(taskService)
	fileHandler := handlers.NewFileHandler(fileService, "uploads")
	sseHandler := handlers.NewSSEHandler(taskService)
	webHandler := handlers.NewWebHandler()

	// 5. 创建 Gin 路由器
	router := gin.Default()

	// 6. 应用中间件
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// 7. 静态文件服务
	router.Static("/static", "./web/static")

	// 8. 加载 HTML 模板
	router.LoadHTMLGlob("web/templates/*")

	// 9. Web 页面路由
	router.GET("/", webHandler.HomePage)
	router.GET("/tasks", webHandler.TasksPage)
	router.GET("/files", webHandler.FilesPage)
	router.GET("/monitor", webHandler.MonitorPage)

	// 10. API 路由组
	api := router.Group("/api")
	{
		// 任务管理 API
		tasks := api.Group("/tasks")
		{
			tasks.GET("", taskHandler.GetTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.POST("", taskHandler.CreateTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)

			// 异步处理任务（提交到 Worker Pool）
			tasks.POST("/:id/process", func(c *gin.Context) {
				taskID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
					return
				}

				// 获取任务信息
				task, err := taskService.GetByID(taskID)
				if err != nil || task == nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
					return
				}

				// 提交到 Worker Pool
				job := services.TaskJob{
					ID:          task.ID,
					Title:       task.Title,
					Description: task.Description,
				}

				if err := workerPool.Submit(job); err != nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusAccepted, gin.H{
					"message": "任务已提交到 Worker Pool 进行异步处理",
					"task_id": taskID,
				})
			})
		}

		// 文件管理 API
		files := api.Group("/files")
		{
			files.GET("", fileHandler.GetFiles)
			files.POST("/upload", fileHandler.UploadFile)
			files.GET("/:id/download", fileHandler.DownloadFile)
			files.DELETE("/:id", fileHandler.DeleteFile)
		}

		// SSE 实时推送 API
		sse := api.Group("/sse")
		{
			sse.GET("/tasks", sseHandler.SSETaskUpdates)
			sse.GET("/system", sseHandler.SSESystemStats)
		}
	}

	// 11. 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 12. 在 goroutine 中启动服务器
	go func() {
		log.Println("======================================")
		log.Println("服务器启动成功！")
		log.Println("访问地址: http://localhost:8080")
		log.Println("======================================")
		log.Println("")
		log.Println("可用页面:")
		log.Println("  主页:     http://localhost:8080/")
		log.Println("  任务管理: http://localhost:8080/tasks")
		log.Println("  文件管理: http://localhost:8080/files")
		log.Println("  实时监控: http://localhost:8080/monitor")
		log.Println("")
		log.Println("API 端点:")
		log.Println("  GET    /api/tasks")
		log.Println("  POST   /api/tasks")
		log.Println("  GET    /api/files")
		log.Println("  POST   /api/files/upload")
		log.Println("  GET    /api/sse/system")
		log.Println("  GET    /api/sse/tasks")
		log.Println("======================================")
		log.Println("按 Ctrl+C 停止服务器")
		log.Println("")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 13. 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("")
	log.Println("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	log.Println("服务器已安全关闭")
}
