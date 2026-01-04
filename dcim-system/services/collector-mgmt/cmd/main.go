package main

import (
	"flag"
	"fmt"

	"github.com/dcim/services/collector-mgmt/internal/handler"
	"github.com/dcim/services/collector-mgmt/internal/service"
	"github.com/dcim/services/collector-mgmt/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// 连接Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 创建服务
	taskService := service.NewTaskService(redisClient)

	// 创建处理器
	taskHandler := handler.NewTaskHandler(taskService)

	// 初始化Gin
	gin.SetMode(cfg.Server.Mode)
	router := gin.Default()

	// 注册路由
	taskHandler.RegisterRoutes(router)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	fmt.Printf("采集管理服务启动在 %s\n", addr)
	if err := router.Run(addr); err != nil {
		panic(err)
	}
}
