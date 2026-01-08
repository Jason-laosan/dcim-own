package database

import (
	"log"
)

// Task GORM 模型定义（用于迁移）
type Task struct {
	ID          int     `gorm:"primaryKey;autoIncrement"`
	Title       string  `gorm:"not null"`
	Description string  `gorm:"type:text"`
	Status      string  `gorm:"default:'pending';index"`
	Priority    int     `gorm:"default:1"`
	CreatedAt   string  `gorm:"autoCreateTime;index"`
	UpdatedAt   string  `gorm:"autoUpdateTime"`
	CompletedAt *string `gorm:"type:datetime"`
}

// File GORM 模型定义（用于迁移）
type File struct {
	ID           int    `gorm:"primaryKey;autoIncrement"`
	OriginalName string `gorm:"not null"`
	StoredName   string `gorm:"not null"`
	FileSize     int64  `gorm:"not null"`
	MimeType     string
	UploadPath   string `gorm:"not null"`
	UploadedAt   string `gorm:"autoCreateTime;index"`
}

// User GORM 模型定义（用于迁移）
type User struct {
	ID        int    `gorm:"primaryKey;autoIncrement"`
	Username  string `gorm:"not null;uniqueIndex"`
	Email     string `gorm:"not null;uniqueIndex"`
	Password  string `gorm:"not null"`
	CreatedAt string `gorm:"autoCreateTime"`
	UpdatedAt string `gorm:"autoUpdateTime"`
}

// RunMigrations 执行数据库迁移，创建所有必要的表
func RunMigrations() error {
	// 使用 GORM AutoMigrate 自动创建表和索引
	err := DB.AutoMigrate(&Task{}, &File{}, &User{})
	if err != nil {
		return err
	}

	log.Println("数据库表已创建或更新")
	return nil
}
