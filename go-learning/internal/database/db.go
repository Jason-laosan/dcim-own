package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB 全局数据库连接实例
var DB *gorm.DB

// InitDB 初始化数据库连接
// dbPath: 数据库文件路径（例如: "tasks.db"）
func InitDB(dbPath string) error {
	var err error

	// 打开数据库连接
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 获取底层的 sql.DB 以设置连接池参数
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(25)   // 最大打开连接数
	sqlDB.SetMaxIdleConns(5)    // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(0) // 连接最大生命周期（0 表示不限制）

	log.Println("数据库连接成功")

	// 执行数据库迁移（创建表）
	if err = RunMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		log.Println("关闭数据库连接")
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB 获取数据库连接实例
func GetDB() *gorm.DB {
	return DB
}
