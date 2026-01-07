package protocol

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"

	"opc-collector/pkg/config"
	"opc-collector/pkg/logger"
	"opc-collector/pkg/models"
)

var (
	// ErrPoolExhausted is returned when the connection pool is at max capacity
	ErrPoolExhausted = errors.New("connection pool exhausted")
	// ErrPoolClosed is returned when attempting to use a closed pool
	ErrPoolClosed = errors.New("connection pool closed")
)

// PooledConnection represents a connection from the pool
type PooledConnection struct {
	conn       interface{} // The actual connection (OPC UA Client, etc.)
	deviceID   string
	createdAt  time.Time
	lastUsedAt time.Time
	inUse      bool
	mu         sync.Mutex
}

// isExpired checks if the connection has expired
func (pc *PooledConnection) isExpired(maxLifetime, idleTimeout time.Duration) bool {
	now := time.Now()

	// Check max lifetime
	if maxLifetime > 0 && now.Sub(pc.createdAt) > maxLifetime {
		return true
	}

	// Check idle timeout
	if idleTimeout > 0 && !pc.inUse && now.Sub(pc.lastUsedAt) > idleTimeout {
		return true
	}

	return false
}

// Close closes the connection
func (pc *PooledConnection) Close() error {
	if closer, ok := pc.conn.(interface{ Close() error }); ok {
		return closer.Close()
	}
	return nil
}

// ConnectionPool manages a pool of connections to OPC servers
type ConnectionPool struct {
	protocol    models.OPCProtocol
	maxIdle     int
	maxOpen     int
	maxLifetime time.Duration
	idleTimeout time.Duration

	mu          sync.RWMutex
	connections map[string]*PooledConnection // key: deviceID
	idle        []*PooledConnection
	active      int
	closed      bool

	logger *zap.Logger
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(protocol models.OPCProtocol, cfg config.PoolConfig) *ConnectionPool {
	return &ConnectionPool{
		protocol:    protocol,
		maxIdle:     cfg.MaxIdle,
		maxOpen:     cfg.MaxOpen,
		maxLifetime: time.Duration(cfg.MaxLifetime) * time.Second,
		idleTimeout: time.Duration(cfg.IdleTimeout) * time.Second,
		connections: make(map[string]*PooledConnection),
		idle:        make([]*PooledConnection, 0, cfg.MaxIdle),
		logger:      logger.Named(string(protocol) + "-pool"),
	}
}

// Get retrieves a connection from the pool or creates a new one
func (p *ConnectionPool) Get(device *models.OPCDevice, createFunc func(*models.OPCDevice) (interface{}, error)) (*PooledConnection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, ErrPoolClosed
	}

	// Clean up expired connections in idle pool
	p.cleanupExpired()

	// Try to find an idle connection for this device
	for i, conn := range p.idle {
		if conn.deviceID == device.ID && !conn.isExpired(p.maxLifetime, p.idleTimeout) {
			// Remove from idle list
			p.idle = append(p.idle[:i], p.idle[i+1:]...)

			// Mark as in use
			conn.mu.Lock()
			conn.inUse = true
			conn.lastUsedAt = time.Now()
			conn.mu.Unlock()

			p.active++
			p.logger.Debug("reused connection from pool",
				zap.String("device_id", device.ID),
				zap.Int("active", p.active),
				zap.Int("idle", len(p.idle)))

			return conn, nil
		}
	}

	// Check if we can create a new connection
	if p.active >= p.maxOpen {
		p.logger.Warn("connection pool exhausted",
			zap.Int("active", p.active),
			zap.Int("max", p.maxOpen))
		return nil, ErrPoolExhausted
	}

	// Create new connection
	rawConn, err := createFunc(device)
	if err != nil {
		return nil, err
	}

	conn := &PooledConnection{
		conn:       rawConn,
		deviceID:   device.ID,
		createdAt:  time.Now(),
		lastUsedAt: time.Now(),
		inUse:      true,
	}

	p.connections[device.ID] = conn
	p.active++

	p.logger.Debug("created new connection",
		zap.String("device_id", device.ID),
		zap.Int("active", p.active),
		zap.Int("idle", len(p.idle)))

	return conn, nil
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(conn *PooledConnection) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		conn.Close()
		return
	}

	conn.mu.Lock()
	conn.inUse = false
	conn.lastUsedAt = time.Now()
	conn.mu.Unlock()

	// Check if connection is expired
	if conn.isExpired(p.maxLifetime, p.idleTimeout) {
		conn.Close()
		delete(p.connections, conn.deviceID)
		p.active--
		p.logger.Debug("closed expired connection",
			zap.String("device_id", conn.deviceID),
			zap.Int("active", p.active))
		return
	}

	// Add to idle pool if not full
	if len(p.idle) < p.maxIdle {
		p.idle = append(p.idle, conn)
		p.active--
		p.logger.Debug("returned connection to pool",
			zap.String("device_id", conn.deviceID),
			zap.Int("active", p.active),
			zap.Int("idle", len(p.idle)))
	} else {
		// Pool is full, close the connection
		conn.Close()
		delete(p.connections, conn.deviceID)
		p.active--
		p.logger.Debug("pool full, closed connection",
			zap.String("device_id", conn.deviceID),
			zap.Int("active", p.active))
	}
}

// cleanupExpired removes expired connections from the idle pool
func (p *ConnectionPool) cleanupExpired() {
	var validConns []*PooledConnection

	for _, conn := range p.idle {
		if conn.isExpired(p.maxLifetime, p.idleTimeout) {
			conn.Close()
			delete(p.connections, conn.deviceID)
			p.logger.Debug("cleaned up expired connection",
				zap.String("device_id", conn.deviceID))
		} else {
			validConns = append(validConns, conn)
		}
	}

	p.idle = validConns
}

// Stats returns current pool statistics
func (p *ConnectionPool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		Active:  p.active,
		Idle:    len(p.idle),
		Total:   len(p.connections),
		MaxOpen: p.maxOpen,
		MaxIdle: p.maxIdle,
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}

	p.closed = true

	// Close all connections
	for _, conn := range p.connections {
		conn.Close()
	}

	// Clear maps and slices
	p.connections = make(map[string]*PooledConnection)
	p.idle = nil
	p.active = 0

	p.logger.Info("connection pool closed")

	return nil
}

// PoolStats contains connection pool statistics
type PoolStats struct {
	Active  int
	Idle    int
	Total   int
	MaxOpen int
	MaxIdle int
}
