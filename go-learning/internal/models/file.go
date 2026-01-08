package models

import (
	"time"

	"gorm.io/gorm"
)

// File 文件模型结构体
type File struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	OriginalName string    `json:"original_name" gorm:"not null"`
	StoredName   string    `json:"stored_name" gorm:"not null"`
	FileSize     int64     `json:"file_size" gorm:"not null"`
	MimeType     string    `json:"mime_type"`
	UploadPath   string    `json:"upload_path" gorm:"not null"`
	UploadedAt   time.Time `json:"uploaded_at" gorm:"autoCreateTime;index"`
}

// FileService 文件服务，封装数据库操作
type FileService struct {
	DB *gorm.DB
}

// NewFileService 创建文件服务实例
func NewFileService(db *gorm.DB) *FileService {
	return &FileService{DB: db}
}

// GetAll 获取所有文件记录
func (s *FileService) GetAll() ([]File, error) {
	var files []File
	err := s.DB.Order("uploaded_at DESC").Find(&files).Error
	return files, err
}

// GetByID 根据 ID 获取文件记录
func (s *FileService) GetByID(id int) (*File, error) {
	var file File
	err := s.DB.First(&file, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 文件不存在
	}
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// Create 创建文件记录
func (s *FileService) Create(file *File) error {
	return s.DB.Create(file).Error
}

// Delete 删除文件记录
func (s *FileService) Delete(id int) error {
	return s.DB.Delete(&File{}, id).Error
}
