package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dcim/collector-agent/internal/protocol"
	"github.com/dgraph-io/badger/v4"
)

// LocalCache 本地缓存
type LocalCache struct {
	db            *badger.DB
	maxCacheTime  time.Duration // 最大缓存时长
	cleanInterval time.Duration // 清理间隔
}

// NewLocalCache 创建本地缓存实例
func NewLocalCache(path string, maxCacheTimeHours int, cleanIntervalMinutes int) (*LocalCache, error) {
	opts := badger.DefaultOptions(path)
	opts.Logger = nil // 关闭badger默认日志

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger: %w", err)
	}

	cache := &LocalCache{
		db:            db,
		maxCacheTime:  time.Duration(maxCacheTimeHours) * time.Hour,
		cleanInterval: time.Duration(cleanIntervalMinutes) * time.Minute,
	}

	// 启动定期清理任务
	go cache.startCleanTask()

	return cache, nil
}

// Save 保存数据到缓存
func (c *LocalCache) Save(data *protocol.DeviceData) error {
	key := c.generateKey(data.DeviceID, data.Timestamp)

	value, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = c.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value).WithTTL(c.maxCacheTime)
		return txn.SetEntry(entry)
	})

	return err
}

// GetAll 获取所有缓存数据
func (c *LocalCache) GetAll() ([]*protocol.DeviceData, error) {
	var dataList []*protocol.DeviceData

	err := c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var data protocol.DeviceData
				if err := json.Unmarshal(val, &data); err != nil {
					return err
				}
				dataList = append(dataList, &data)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	return dataList, err
}

// Delete 删除缓存数据
func (c *LocalCache) Delete(deviceID string, timestamp time.Time) error {
	key := c.generateKey(deviceID, timestamp)

	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Count 获取缓存数据数量
func (c *LocalCache) Count() (int, error) {
	count := 0

	err := c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return nil
	})

	return count, err
}

// Close 关闭缓存
func (c *LocalCache) Close() error {
	return c.db.Close()
}

// generateKey 生成缓存key
func (c *LocalCache) generateKey(deviceID string, timestamp time.Time) string {
	return fmt.Sprintf("%s_%d", deviceID, timestamp.Unix())
}

// startCleanTask 启动定期清理过期数据任务
func (c *LocalCache) startCleanTask() {
	ticker := time.NewTicker(c.cleanInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.db.RunValueLogGC(0.5)
	}
}
