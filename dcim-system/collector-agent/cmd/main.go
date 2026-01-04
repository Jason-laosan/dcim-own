package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/dcim/collector-agent/internal/agent"
	"github.com/dcim/collector-agent/pkg/config"
	"github.com/dcim/collector-agent/pkg/logger"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 初始化日志
	if err := logger.InitLogger(); err != nil {
		panic(err)
	}
	defer logger.Sync()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Log.Fatal("failed to load config", zap.Error(err))
	}

	logger.Log.Info("config loaded",
		zap.String("agent_id", cfg.Agent.ID),
		zap.String("agent_name", cfg.Agent.Name))

	// 创建Agent
	agentInstance, err := agent.NewAgent(cfg)
	if err != nil {
		logger.Log.Fatal("failed to create agent", zap.Error(err))
	}

	// 启动Agent
	if err := agentInstance.Start(); err != nil {
		logger.Log.Fatal("failed to start agent", zap.Error(err))
	}

	// 监听退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Log.Info("agent is running, press Ctrl+C to exit")

	// 等待退出信号
	<-sigChan

	// 优雅退出
	logger.Log.Info("shutting down agent...")
	agentInstance.Stop()
	logger.Log.Info("agent shutdown complete")
}
