package protocol

import (
	"context"

	"opc-collector/pkg/models"
)

// Protocol defines the interface that all OPC protocol implementations must satisfy
type Protocol interface {
	// Name returns the protocol name
	Name() string

	// Collect collects metrics from the device
	Collect(ctx context.Context, device *models.OPCDevice) (*models.MetricData, error)

	// Close closes the protocol and releases resources
	Close() error
}

// ProtocolFactory creates protocol instances
type ProtocolFactory interface {
	Create(protocol models.OPCProtocol) (Protocol, error)
}
