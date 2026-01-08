# GORM 迁移文档

## 概述

项目已成功从原生 SQL 操作迁移到 **GORM** (Go Object-Relational Mapping) 框架。

## 迁移内容

### 1. 依赖变更

**移除的依赖**：
- `modernc.org/sqlite` - 纯 Go SQLite 驱动

**新增的依赖**：
- `gorm.io/gorm` - GORM 核心库
- `gorm.io/driver/sqlite` - GORM SQLite 驱动

### 2. 数据库初始化

**文件**: `internal/database/db.go`

**变更**：
- 全局变量从 `*sql.DB` 改为 `*gorm.DB`
- 使用 `gorm.Open()` 替代 `sql.Open()`
- 连接池配置通过 `DB.DB()` 获取底层 `sql.DB` 进行设置

**示例**：
```go
// 旧方式
DB, err = sql.Open("sqlite", dbPath)

// 新方式
DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
```

### 3. 数据库迁移

**文件**: `internal/database/migrations.go`

**变更**：
- 移除了所有手动 SQL 建表语句
- 使用 GORM 的 `AutoMigrate` 功能自动创建表和索引
- 在迁移文件中定义了临时模型结构用于迁移

**示例**：
```go
// 旧方式
createTasksTable := `CREATE TABLE IF NOT EXISTS tasks (...)`
DB.Exec(createTasksTable)

// 新方式
DB.AutoMigrate(&Task{}, &File{}, &User{})
```

### 4. 模型定义

所有模型都添加了 GORM 标签：

#### Task 模型 (`internal/models/task.go`)
```go
type Task struct {
    ID          int        `json:"id" gorm:"primaryKey;autoIncrement"`
    Title       string     `json:"title" gorm:"not null"`
    Description string     `json:"description" gorm:"type:text"`
    Status      string     `json:"status" gorm:"default:'pending';index"`
    Priority    int        `json:"priority" gorm:"default:1"`
    CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime;index"`
    UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
    CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"type:datetime"`
}
```

#### File 模型 (`internal/models/file.go`)
```go
type File struct {
    ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
    OriginalName string    `json:"original_name" gorm:"not null"`
    StoredName   string    `json:"stored_name" gorm:"not null"`
    FileSize     int64     `json:"file_size" gorm:"not null"`
    MimeType     string    `json:"mime_type"`
    UploadPath   string    `json:"upload_path" gorm:"not null"`
    UploadedAt   time.Time `json:"uploaded_at" gorm:"autoCreateTime;index"`
}
```

#### User 模型 (`internal/models/user.go`)
```go
type User struct {
    ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
    Username  string    `json:"username" gorm:"not null;uniqueIndex"`
    Email     string    `json:"email" gorm:"not null;uniqueIndex"`
    Password  string    `json:"-" gorm:"not null"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
```

### 5. Service 层变更

所有 Service 的数据库操作都已简化：

#### TaskService 示例

**查询所有任务**：
```go
// 旧方式
rows, err := s.DB.Query("SELECT ... FROM tasks ORDER BY created_at DESC")
// ... 手动扫描行

// 新方式
var tasks []Task
err := s.DB.Order("created_at DESC").Find(&tasks).Error
```

**根据 ID 查询**：
```go
// 旧方式
err := s.DB.QueryRow("SELECT ... FROM tasks WHERE id = ?", id).Scan(...)

// 新方式
var task Task
err := s.DB.First(&task, id).Error
```

**创建记录**：
```go
// 旧方式
result, err := s.DB.Exec("INSERT INTO tasks (...) VALUES (...)", ...)
id, _ := result.LastInsertId()
task.ID = int(id)

// 新方式
err := s.DB.Create(task).Error
// ID 自动填充到 task.ID
```

**更新记录**：
```go
// 旧方式
_, err := s.DB.Exec("UPDATE tasks SET ... WHERE id = ?", ...)

// 新方式
err := s.DB.Save(task).Error
```

**删除记录**：
```go
// 旧方式
_, err := s.DB.Exec("DELETE FROM tasks WHERE id = ?", id)

// 新方式
err := s.DB.Delete(&Task{}, id).Error
```

**条件查询**：
```go
// 旧方式
rows, err := s.DB.Query("SELECT ... FROM tasks WHERE status = ?", status)

// 新方式
var tasks []Task
err := s.DB.Where("status = ?", status).Find(&tasks).Error
```

## GORM 优势

### 1. **代码简洁**
- 减少了大量样板代码
- 无需手动编写 SQL 语句
- 自动处理结果扫描

### 2. **类型安全**
- 编译时检查
- 减少 SQL 注入风险
- 自动参数绑定

### 3. **自动迁移**
- `AutoMigrate` 自动创建表和索引
- 支持表结构变更
- 无需手动维护 SQL 迁移脚本

### 4. **关联关系**
- 支持一对一、一对多、多对多关系
- 预加载（Preload）功能
- 级联操作

### 5. **钩子函数**
- BeforeCreate, AfterCreate
- BeforeUpdate, AfterUpdate
- BeforeDelete, AfterDelete

### 6. **事务支持**
```go
db.Transaction(func(tx *gorm.DB) error {
    // 在事务中执行操作
    return nil
})
```

### 7. **软删除**
```go
type Model struct {
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

## 常用 GORM 操作

### 查询

```go
// 查询单条记录
db.First(&user, 1)                    // 根据主键查询
db.First(&user, "email = ?", email)   // 根据条件查询

// 查询多条记录
db.Find(&users)                       // 查询所有
db.Where("status = ?", "active").Find(&users)  // 条件查询
db.Limit(10).Offset(20).Find(&users)  // 分页

// 排序
db.Order("created_at DESC").Find(&tasks)

// 选择字段
db.Select("id", "title").Find(&tasks)
```

### 创建

```go
// 创建单条记录
db.Create(&user)

// 批量创建
db.Create(&[]User{user1, user2, user3})
```

### 更新

```go
// 更新所有字段
db.Save(&user)

// 更新指定字段
db.Model(&user).Update("status", "active")
db.Model(&user).Updates(map[string]interface{}{"status": "active", "priority": 5})

// 批量更新
db.Model(&Task{}).Where("status = ?", "pending").Update("status", "processing")
```

### 删除

```go
// 删除单条记录
db.Delete(&user, 1)

// 批量删除
db.Where("status = ?", "inactive").Delete(&User{})
```

### 高级查询

```go
// 原生 SQL
db.Raw("SELECT * FROM tasks WHERE status = ?", "pending").Scan(&tasks)

// 子查询
db.Where("id IN (?)", db.Table("tasks").Select("id").Where("status = ?", "pending")).Find(&tasks)

// 分组
db.Model(&Task{}).Select("status, count(*) as count").Group("status").Scan(&results)
```

## 注意事项

1. **记录未找到错误**
   - GORM 使用 `gorm.ErrRecordNotFound` 而不是 `sql.ErrNoRows`

2. **零值更新**
   - 使用 `Updates` 时，零值字段不会被更新
   - 使用 `Select` 强制更新零值字段

3. **时间字段**
   - `CreatedAt` 和 `UpdatedAt` 会自动管理
   - 使用 `gorm:"autoCreateTime"` 和 `gorm:"autoUpdateTime"` 标签

4. **性能优化**
   - 使用 `Preload` 避免 N+1 查询
   - 使用索引提高查询性能
   - 批量操作使用 `CreateInBatches`

## 数据库兼容性

当前使用 SQLite，GORM 也支持其他数据库：

- MySQL: `gorm.io/driver/mysql`
- PostgreSQL: `gorm.io/driver/postgres`
- SQL Server: `gorm.io/driver/sqlserver`

切换数据库只需更改驱动，模型代码无需修改。

## 迁移完成

✅ 所有模型已迁移到 GORM
✅ 所有 Service 方法已更新
✅ 数据库初始化已更新
✅ 构建测试通过
✅ 功能保持不变

项目现在使用现代化的 ORM 框架，代码更简洁、更易维护！
