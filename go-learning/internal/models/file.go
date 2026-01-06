package models

import (
	"database/sql"
	"time"
)

// File 文件模型结构体
type File struct {
	ID           int       `json:"id"`
	OriginalName string    `json:"original_name"`
	StoredName   string    `json:"stored_name"`
	FileSize     int64     `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	UploadPath   string    `json:"upload_path"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// FileService 文件服务，封装数据库操作
type FileService struct {
	DB *sql.DB
}

// NewFileService 创建文件服务实例
func NewFileService(db *sql.DB) *FileService {
	return &FileService{DB: db}
}

// GetAll 获取所有文件记录
func (s *FileService) GetAll() ([]File, error) {
	query := `SELECT id, original_name, stored_name, file_size, mime_type, upload_path, uploaded_at
	          FROM files ORDER BY uploaded_at DESC`

	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var file File
		err := rows.Scan(
			&file.ID,
			&file.OriginalName,
			&file.StoredName,
			&file.FileSize,
			&file.MimeType,
			&file.UploadPath,
			&file.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, rows.Err()
}

// GetByID 根据 ID 获取文件记录
func (s *FileService) GetByID(id int) (*File, error) {
	query := `SELECT id, original_name, stored_name, file_size, mime_type, upload_path, uploaded_at
	          FROM files WHERE id = ?`

	var file File
	err := s.DB.QueryRow(query, id).Scan(
		&file.ID,
		&file.OriginalName,
		&file.StoredName,
		&file.FileSize,
		&file.MimeType,
		&file.UploadPath,
		&file.UploadedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // 文件不存在
	}
	if err != nil {
		return nil, err
	}

	return &file, nil
}

// Create 创建文件记录
func (s *FileService) Create(file *File) error {
	query := `INSERT INTO files (original_name, stored_name, file_size, mime_type, upload_path)
	          VALUES (?, ?, ?, ?, ?)`

	result, err := s.DB.Exec(query,
		file.OriginalName,
		file.StoredName,
		file.FileSize,
		file.MimeType,
		file.UploadPath,
	)
	if err != nil {
		return err
	}

	// 获取新插入记录的 ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	file.ID = int(id)

	return nil
}

// Delete 删除文件记录
func (s *FileService) Delete(id int) error {
	query := `DELETE FROM files WHERE id = ?`
	_, err := s.DB.Exec(query, id)
	return err
}
