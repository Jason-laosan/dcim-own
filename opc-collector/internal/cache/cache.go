package cache

import (
	"encoding/json"
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"go.uber.org/zap"

	"github.com/yourusername/opc-collector/pkg/config"
	"github.com/yourusername/opc-collector/pkg/logger"
	"github.com/yourusername/opc-collector/pkg/models"
)

// Cache provides local persistent storage using Badger DB
type Cache struct {
	db     *badger.DB
	ttl    time.Duration
	logger *zap.Logger
}

// NewCache creates a new cache instance
func NewCache(cfg config.CacheConfig) (*Cache, error) {
	opts := badger.DefaultOptions(cfg.Path)
	opts.Logger = nil // Disable badger's own logging

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open badger db: %w", err)
	}

	cache := &Cache{
		db:     db,
		ttl:    time.Duration(cfg.TTL) * time.Second,
		logger: logger.Named("cache"),
	}

	// Start garbage collection
	go cache.runGC(time.Duration(cfg.GCInterval) * time.Second)

	return cache, nil
}

// Store stores metric data in the cache
func (c *Cache) Store(data *models.MetricData) error {
	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	// Generate key
	key := []byte(fmt.Sprintf("metric:%s:%d", data.DeviceID, data.Timestamp.UnixNano()))

	// Store with TTL
	err = c.db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, jsonData).WithTTL(c.ttl)
		return txn.SetEntry(e)
	})

	if err != nil {
		return fmt.Errorf("store in cache: %w", err)
	}

	return nil
}

// StoreBatch stores a batch of metrics in the cache
func (c *Cache) StoreBatch(data []*models.MetricData) error {
	wb := c.db.NewWriteBatch()
	defer wb.Cancel()

	for _, metric := range data {
		jsonData, err := json.Marshal(metric)
		if err != nil {
			c.logger.Warn("failed to marshal metric", zap.Error(err))
			continue
		}

		key := []byte(fmt.Sprintf("metric:%s:%d", metric.DeviceID, metric.Timestamp.UnixNano()))

		e := badger.NewEntry(key, jsonData).WithTTL(c.ttl)
		if err := wb.SetEntry(e); err != nil {
			c.logger.Warn("failed to add to batch", zap.Error(err))
		}
	}

	if err := wb.Flush(); err != nil {
		return fmt.Errorf("flush batch: %w", err)
	}

	c.logger.Debug("stored batch in cache", zap.Int("count", len(data)))

	return nil
}

// GetAll retrieves all cached metrics
func (c *Cache) GetAll() ([]*models.MetricData, error) {
	var metrics []*models.MetricData

	err := c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true

		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("metric:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			err := item.Value(func(val []byte) error {
				var metric models.MetricData
				if err := json.Unmarshal(val, &metric); err != nil {
					c.logger.Warn("failed to unmarshal metric", zap.Error(err))
					return nil // Continue iteration
				}
				metrics = append(metrics, &metric)
				return nil
			})

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("iterate cache: %w", err)
	}

	return metrics, nil
}

// DeleteOlderThan deletes metrics older than the specified duration
func (c *Cache) DeleteOlderThan(duration time.Duration) error {
	cutoff := time.Now().Add(-duration)

	err := c.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("metric:")
		var keysToDelete [][]byte

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()

			// Check expiration
			if item.ExpiresAt() > 0 && item.ExpiresAt() < uint64(cutoff.Unix()) {
				keysToDelete = append(keysToDelete, item.KeyCopy(nil))
			}
		}

		// Delete keys
		for _, key := range keysToDelete {
			if err := txn.Delete(key); err != nil {
				return err
			}
		}

		c.logger.Debug("deleted old cache entries", zap.Int("count", len(keysToDelete)))

		return nil
	})

	return err
}

// runGC runs periodic garbage collection
func (c *Cache) runGC(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		c.logger.Debug("running garbage collection")

		// Run badger GC
		for {
			err := c.db.RunValueLogGC(0.5)
			if err != nil {
				break
			}
		}
	}
}

// Size returns the current cache size
func (c *Cache) Size() (int64, error) {
	lsm, vlog := c.db.Size()
	return lsm + vlog, nil
}

// Close closes the cache
func (c *Cache) Close() error {
	c.logger.Info("closing cache")
	return c.db.Close()
}
