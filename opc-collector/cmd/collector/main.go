package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/yourusername/opc-collector/internal/agent"
	"github.com/yourusername/opc-collector/pkg/config"
	"github.com/yourusername/opc-collector/pkg/logger"
	"github.com/yourusername/opc-collector/pkg/models"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "configs/config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("OPC Collector\n")
		fmt.Printf("  Version:    %s\n", version)
		fmt.Printf("  Build Time: %s\n", buildTime)
		fmt.Printf("  Git Commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	if err := logger.Init(cfg.Logging); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("OPC Collector starting",
		zap.String("version", version),
		zap.String("build_time", buildTime),
		zap.String("git_commit", gitCommit),
		zap.String("agent_id", cfg.Agent.ID))

	// Load devices
	var devices *config.DeviceList
	switch cfg.Devices.Source {
	case "file":
		devices, err = config.LoadDevices(cfg.Devices.File.Path)
		if err != nil {
			logger.Fatal("Failed to load devices from file",
				zap.String("path", cfg.Devices.File.Path),
				zap.Error(err))
		}
		logger.Info("Loaded devices from file",
			zap.String("path", cfg.Devices.File.Path),
			zap.Int("count", len(devices.Devices)))

	case "database":
		// TODO: Implement database device loader
		logger.Fatal("Database device source not yet implemented")

	case "etcd":
		// TODO: Implement etcd device loader
		logger.Fatal("Etcd device source not yet implemented")

	default:
		logger.Fatal("Invalid device source", zap.String("source", cfg.Devices.Source))
	}

	// Validate device count
	if len(devices.Devices) == 0 {
		logger.Fatal("No devices configured")
	}

	if len(devices.Devices) > cfg.Agent.MaxDevices {
		logger.Warn("Device count exceeds max_devices limit",
			zap.Int("device_count", len(devices.Devices)),
			zap.Int("max_devices", cfg.Agent.MaxDevices))
	}

	// Convert devices to pointers
	devicePointers := make([]*models.OPCDevice, len(devices.Devices))
	for i := range devices.Devices {
		devicePointers[i] = &devices.Devices[i]
	}

	// Create agent
	ag, err := agent.NewAgent(cfg, devicePointers)
	if err != nil {
		logger.Fatal("Failed to create agent", zap.Error(err))
	}

	// Start agent
	if err := ag.Start(); err != nil {
		logger.Fatal("Failed to start agent", zap.Error(err))
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("OPC Collector running, press Ctrl+C to stop")

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	// Graceful shutdown
	ag.Stop()

	logger.Info("OPC Collector stopped")
}
