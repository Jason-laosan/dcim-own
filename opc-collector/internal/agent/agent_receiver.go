package agent

import (
	"fmt"

	"go.uber.org/zap"

	"opc-collector/internal/monitor"
	"opc-collector/internal/receiver"
	"opc-collector/pkg/models"
)

// initializeReceiver initializes the push mode receiver
func (a *Agent) initializeReceiver() error {
	// Skip if receiver is disabled
	if !a.config.Receiver.Enabled {
		a.logger.Info("receiver is disabled, operating in pull mode only")
		return nil
	}

	// Create receiver manager with data handler
	receiverManager, err := receiver.NewManager(&a.config.Receiver, a.handlePushData)
	if err != nil {
		return fmt.Errorf("create receiver manager: %w", err)
	}

	a.receiver = receiverManager

	a.logger.Info("receiver initialized",
		zap.Bool("http_enabled", a.config.Receiver.HTTP.Enabled),
		zap.Bool("mqtt_enabled", a.config.Receiver.MQTT.Enabled))

	return nil
}

// handlePushData handles data pushed from downstream devices/systems
// This is the callback function invoked when data is received via HTTP or MQTT
func (a *Agent) handlePushData(data *models.MetricData) error {
	a.logger.Debug("received push data",
		zap.String("device_id", data.DeviceID),
		zap.String("device_ip", data.DeviceIP),
		zap.Int("metrics_count", len(data.Metrics)))

	// Validate data
	if data.DeviceID == "" {
		a.logger.Warn("received push data without device_id")
		monitor.ReceiverErrorsTotal.WithLabelValues("validation_error").Inc()
		return fmt.Errorf("device_id is required")
	}

	// Update receiver metrics
	monitor.ReceiverDataReceived.WithLabelValues(data.DeviceID, "push").Inc()

	// Add to batcher (same processing pipeline as pull mode)
	a.batcher.Add(data)

	// Update batch metrics
	monitor.BatchItemsReceived.Inc()

	a.logger.Debug("push data added to batcher",
		zap.String("device_id", data.DeviceID))

	return nil
}

// startReceiver starts the receiver manager
func (a *Agent) startReceiver() error {
	if a.receiver == nil {
		return nil
	}

	if err := a.receiver.Start(); err != nil {
		return fmt.Errorf("start receiver: %w", err)
	}

	a.logger.Info("receiver started",
		zap.Int("active_receivers", a.receiver.GetActiveReceivers()))

	return nil
}

// stopReceiver stops the receiver manager
func (a *Agent) stopReceiver() {
	if a.receiver != nil {
		a.logger.Info("stopping receiver")
		a.receiver.Stop()
	}
}
