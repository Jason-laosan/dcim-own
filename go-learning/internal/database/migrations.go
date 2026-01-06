package database

import "log"

// RunMigrations 执行数据库迁移，创建所有必要的表
func RunMigrations() error {
	// 创建 tasks 表
	createTasksTable := `
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT DEFAULT 'pending',
		priority INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		completed_at DATETIME
	);
	`

	if _, err := DB.Exec(createTasksTable); err != nil {
		return err
	}
	log.Println("tasks 表已创建或已存在")

	// 创建 files 表
	createFilesTable := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		original_name TEXT NOT NULL,
		stored_name TEXT NOT NULL,
		file_size INTEGER NOT NULL,
		mime_type TEXT,
		upload_path TEXT NOT NULL,
		uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := DB.Exec(createFilesTable); err != nil {
		return err
	}
	log.Println("files 表已创建或已存在")

	// 创建索引以提高查询性能
	createIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);",
		"CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_files_uploaded_at ON files(uploaded_at);",
	}

	for _, indexSQL := range createIndexes {
		if _, err := DB.Exec(indexSQL); err != nil {
			return err
		}
	}
	log.Println("数据库索引已创建")

	return nil
}
