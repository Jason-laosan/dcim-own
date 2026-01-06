package handlers

import (
	"fmt"
	"go-learning/internal/models"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// FileHandler 文件处理器
type FileHandler struct {
	fileService *models.FileService
	uploadDir   string
}

// NewFileHandler 创建文件处理器实例
func NewFileHandler(fileService *models.FileService, uploadDir string) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		uploadDir:   uploadDir,
	}
}

// GetFiles 获取所有文件列表
// GET /api/files
func (h *FileHandler) GetFiles(c *gin.Context) {
	files, err := h.fileService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch files",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, files)
}

// UploadFile 上传文件
// POST /api/files/upload
func (h *FileHandler) UploadFile(c *gin.Context) {
	// 从表单中获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file uploaded",
			"details": err.Error(),
		})
		return
	}

	// 生成唯一的文件名（使用 UUID）
	ext := filepath.Ext(file.Filename)
	storedName := uuid.New().String() + ext
	uploadPath := filepath.Join(h.uploadDir, storedName)

	// 保存文件到磁盘
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
			"details": err.Error(),
		})
		return
	}

	// 获取 MIME 类型
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream" // 默认类型
	}

	// 创建文件记录
	fileRecord := &models.File{
		OriginalName: file.Filename,
		StoredName:   storedName,
		FileSize:     file.Size,
		MimeType:     mimeType,
		UploadPath:   uploadPath,
	}

	if err := h.fileService.Create(fileRecord); err != nil {
		// 如果数据库操作失败，删除已上传的文件
		os.Remove(uploadPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file record",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"file": fileRecord,
	})
}

// DownloadFile 下载文件
// GET /api/files/:id/download
func (h *FileHandler) DownloadFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file ID",
		})
		return
	}

	// 从数据库获取文件记录
	file, err := h.fileService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch file record",
			"details": err.Error(),
		})
		return
	}

	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(file.UploadPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found on disk",
		})
		return
	}

	// 设置下载响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.OriginalName))
	c.Header("Content-Type", file.MimeType)

	// 返回文件
	c.File(file.UploadPath)
}

// DeleteFile 删除文件
// DELETE /api/files/:id
func (h *FileHandler) DeleteFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file ID",
		})
		return
	}

	// 从数据库获取文件记录
	file, err := h.fileService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch file record",
			"details": err.Error(),
		})
		return
	}

	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	// 从数据库删除记录
	if err := h.fileService.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file record",
			"details": err.Error(),
		})
		return
	}

	// 从磁盘删除文件
	if err := os.Remove(file.UploadPath); err != nil {
		// 即使文件删除失败，也不返回错误（文件可能已被手动删除）
		fmt.Printf("Warning: failed to delete file from disk: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}
