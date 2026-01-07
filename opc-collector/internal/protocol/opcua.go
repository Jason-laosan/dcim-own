package protocol

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"go.uber.org/zap"

	"github.com/yourusername/opc-collector/pkg/config"
	"github.com/yourusername/opc-collector/pkg/logger"
	"github.com/yourusername/opc-collector/pkg/models"
)

// OPCUAProtocol implements the Protocol interface for OPC UA
type OPCUAProtocol struct {
	pool   *ConnectionPool
	logger *zap.Logger
}

// NewOPCUAProtocol creates a new OPC UA protocol handler
func NewOPCUAProtocol(poolCfg config.PoolConfig) *OPCUAProtocol {
	return &OPCUAProtocol{
		pool:   NewConnectionPool(models.ProtocolOPCUA, poolCfg),
		logger: logger.Named("opcua"),
	}
}

// Name returns the protocol name
func (p *OPCUAProtocol) Name() string {
	return "OPC UA"
}

// Collect collects metrics from an OPC UA device
func (p *OPCUAProtocol) Collect(ctx context.Context, device *models.OPCDevice) (*models.MetricData, error) {
	startTime := time.Now()

	// Get connection from pool
	conn, err := p.pool.Get(device, p.createConnection)
	if err != nil {
		p.logger.Error("failed to get connection",
			zap.String("device_id", device.ID),
			zap.Error(err))
		return nil, fmt.Errorf("get connection: %w", err)
	}
	defer p.pool.Put(conn)

	client := conn.conn.(*opcua.Client)

	// Build read request
	nodeIDs := make([]*ua.ReadValueID, 0, len(device.Metrics))
	metricMap := make(map[int]*models.MetricDefinition)

	for i, metric := range device.Metrics {
		nodeID, err := ua.ParseNodeID(metric.NodeID)
		if err != nil {
			p.logger.Warn("invalid node ID",
				zap.String("device_id", device.ID),
				zap.String("node_id", metric.NodeID),
				zap.Error(err))
			continue
		}

		nodeIDs = append(nodeIDs, &ua.ReadValueID{
			NodeID: nodeID,
		})
		metricMap[len(nodeIDs)-1] = &device.Metrics[i]
	}

	if len(nodeIDs) == 0 {
		return nil, fmt.Errorf("no valid node IDs to read")
	}

	// Read values
	req := &ua.ReadRequest{
		MaxAge:             2000,
		NodesToRead:        nodeIDs,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	resp, err := client.Read(ctx, req)
	if err != nil {
		p.logger.Error("failed to read values",
			zap.String("device_id", device.ID),
			zap.Error(err))
		return nil, fmt.Errorf("read values: %w", err)
	}

	// Parse results
	metrics := make(map[string]models.MetricValue)
	goodCount := 0

	for i, result := range resp.Results {
		metricDef := metricMap[i]

		if result.Status != ua.StatusOK {
			p.logger.Warn("bad metric quality",
				zap.String("device_id", device.ID),
				zap.String("metric", metricDef.Name),
				zap.Uint32("status", uint32(result.Status)))
			continue
		}

		value := p.parseValue(result.Value, metricDef)

		metrics[metricDef.Name] = models.MetricValue{
			Name:    metricDef.Name,
			Value:   value,
			Unit:    metricDef.Unit,
			Quality: p.qualityToString(result.Status),
		}

		goodCount++
	}

	// Determine overall quality
	quality := models.QualityGood
	if goodCount == 0 {
		quality = models.QualityBad
	} else if goodCount < len(device.Metrics) {
		quality = models.QualityPartial
	}

	duration := time.Since(startTime)
	p.logger.Debug("collected metrics",
		zap.String("device_id", device.ID),
		zap.Int("metrics_count", len(metrics)),
		zap.Duration("duration", duration))

	return &models.MetricData{
		DeviceID:  device.ID,
		DeviceIP:  device.IP,
		Timestamp: time.Now(),
		Metrics:   metrics,
		Tags:      device.Tags,
		Quality:   quality,
	}, nil
}

// createConnection creates a new OPC UA connection
func (p *OPCUAProtocol) createConnection(device *models.OPCDevice) (interface{}, error) {
	endpoint := fmt.Sprintf("opc.tcp://%s:%d", device.IP, device.Port)

	opts := []opcua.Option{
		opcua.SecurityMode(ua.MessageSecurityModeNone),
		opcua.SecurityPolicy("None"),
	}

	// Configure security if specified
	if device.ConnectionConfig.SecurityMode != "" && device.ConnectionConfig.SecurityMode != "None" {
		var secMode ua.MessageSecurityMode
		switch device.ConnectionConfig.SecurityMode {
		case "Sign":
			secMode = ua.MessageSecurityModeSign
		case "SignAndEncrypt":
			secMode = ua.MessageSecurityModeSignAndEncrypt
		default:
			secMode = ua.MessageSecurityModeNone
		}
		opts = append(opts, opcua.SecurityMode(secMode))

		if device.ConnectionConfig.SecurityPolicy != "" {
			opts = append(opts, opcua.SecurityPolicy(device.ConnectionConfig.SecurityPolicy))
		}
	}

	// Configure authentication
	if device.ConnectionConfig.Username != "" {
		opts = append(opts, opcua.AuthUsername(
			device.ConnectionConfig.Username,
			device.ConnectionConfig.Password,
		))
	}

	// Configure certificate if provided
	if device.ConnectionConfig.Certificate != "" {
		opts = append(opts, opcua.CertificateFile(device.ConnectionConfig.Certificate))
	}

	// Create client
	client, err := opcua.NewClient(endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	// Connect with timeout
	timeout := time.Duration(device.ConnectionConfig.Timeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect to %s: %w", endpoint, err)
	}

	p.logger.Info("connected to OPC UA server",
		zap.String("device_id", device.ID),
		zap.String("endpoint", endpoint))

	return client, nil
}

// parseValue parses an OPC UA value according to the metric definition
func (p *OPCUAProtocol) parseValue(value *ua.Variant, metricDef *models.MetricDefinition) interface{} {
	if value == nil || value.Value() == nil {
		return nil
	}

	rawValue := value.Value()

	// Apply scale factor if configured
	if metricDef.ScaleFactor != 0 && metricDef.ScaleFactor != 1.0 {
		switch v := rawValue.(type) {
		case float64:
			return v * metricDef.ScaleFactor
		case float32:
			return float64(v) * metricDef.ScaleFactor
		case int:
			return float64(v) * metricDef.ScaleFactor
		case int64:
			return float64(v) * metricDef.ScaleFactor
		case int32:
			return float64(v) * metricDef.ScaleFactor
		}
	}

	// Convert to appropriate type based on data type definition
	switch metricDef.DataType {
	case "float":
		return p.toFloat64(rawValue)
	case "int":
		return p.toInt64(rawValue)
	case "bool":
		return p.toBool(rawValue)
	case "string":
		return fmt.Sprintf("%v", rawValue)
	default:
		return rawValue
	}
}

// toFloat64 converts a value to float64
func (p *OPCUAProtocol) toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// toInt64 converts a value to int64
func (p *OPCUAProtocol) toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	default:
		return 0
	}
}

// toBool converts a value to bool
func (p *OPCUAProtocol) toBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		b, _ := strconv.ParseBool(val)
		return b
	default:
		return false
	}
}

// qualityToString converts an OPC UA status code to a quality string
func (p *OPCUAProtocol) qualityToString(status ua.StatusCode) string {
	if status == ua.StatusOK {
		return "Good"
	} else if status == ua.StatusUncertain {
		return "Uncertain"
	} else {
		return "Bad"
	}
}

// Close closes the protocol and releases all resources
func (p *OPCUAProtocol) Close() error {
	return p.pool.Close()
}
