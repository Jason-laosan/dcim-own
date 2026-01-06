package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite" // SQLite 驱动（纯 Go 实现，无需 CGO）
)

// DB 全局数据库连接实例
var DB *sql.DB

// InitDB 初始化数据库连接
// dbPath: 数据库文件路径（例如: "tasks.db"）
func InitDB(dbPath string) error {
	var err error

	// 打开数据库连接
	// sqlite 驱动名称，dbPath 是数据库文件路径
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// 测试数据库连接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// 设置连接池参数
	DB.SetMaxOpenConns(25)                // 最大打开连接数
	DB.SetMaxIdleConns(5)                 // 最大空闲连接数
	DB.SetConnMaxLifetime(0)              // 连接最大生命周期（0 表示不限制）

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
		return DB.Close()
	}
	return nil
}

// GetDB 获取数据库连接实例
func GetDB() *sql.DB {
	return DB
}
