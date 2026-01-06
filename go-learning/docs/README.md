# TaskHub - Go 语言学习项目

TaskHub 是一个完整的 Go 语言学习项目，展示了 Go 语言的核心特性和最佳实践，包括数据库操作、文件处理、Server-Sent Events (SSE) 实时推送和协程并发处理。

## 功能特性

### 1. SQLite 数据库操作
- 使用 `database/sql` 标准库
- 完整的 CRUD 操作
- 预编译语句防止 SQL 注入
- 索引优化查询性能

### 2. 文件操作
- 文件上传（multipart/form-data）
- 文件下载（Content-Disposition）
- 文件元数据管理
- UUID 生成唯一文件名

### 3. SSE 实时推送
- 任务状态实时更新
- 系统监控（内存、协程数、GC 次数）
- HTTP 长连接管理
- 自动重连机制

### 4. 协程并发处理
- Worker Pool 模式实现
- Channel 通道通信
- WaitGroup 同步
- 异步任务处理

### 5. Web 界面
- 响应式 HTML/CSS/JavaScript 前端
- 任务管理界面
- 文件管理界面
- 实时监控面板

## 技术栈

- **Go 1.21+** - 编程语言
- **Gin** - Web 框架
- **SQLite** - 嵌入式数据库（modernc.org/sqlite 纯 Go 实现）
- **Server-Sent Events** - 实时通信
- **HTML/CSS/JavaScript** - 前端界面

## 项目结构

```
go-learning/
├── cmd/server/main.go              # 应用入口
├── internal/
│   ├── database/                   # 数据库层
│   │   ├── db.go
│   │   └── migrations.go
│   ├── models/                     # 数据模型
│   │   ├── task.go
│   │   └── file.go
│   ├── handlers/                   # HTTP 处理器
│   │   ├── task_handler.go
│   │   ├── file_handler.go
│   │   ├── sse_handler.go
│   │   └── web_handler.go
│   ├── services/                   # 业务逻辑
│   │   └── worker_pool.go
│   └── middleware/                 # 中间件
│       ├── logger.go
│       └── cors.go
├── web/
│   ├── static/                     # 静态资源
│   │   └── css/style.css
│   └── templates/                  # HTML 模板
│       ├── index.html
│       ├── tasks.html
│       ├── files.html
│       └── monitor.html
├── uploads/                        # 文件上传目录
├── docs/                           # 文档
│   ├── README.md
│   └── API.md
├── go.mod
└── go.sum
```

## 快速开始

### 前置要求

- Go 1.21 或更高版本
- Git（可选）

### 安装步骤

1. **克隆项目**（如果使用 Git）
   ```bash
   cd go-learning
   ```

2. **安装依赖**
   ```bash
   go mod download
   ```

3. **运行项目**
   ```bash
   go run cmd/server/main.go
   ```

4. **访问应用**

   打开浏览器访问：http://localhost:8080

### 可用页面

- **主页**: http://localhost:8080/
- **任务管理**: http://localhost:8080/tasks
- **文件管理**: http://localhost:8080/files
- **实时监控**: http://localhost:8080/monitor

## 使用指南

### 任务管理

1. 访问任务管理页面
2. 填写表单创建新任务
3. 点击"处理"按钮将任务提交到 Worker Pool 异步处理
4. 在实时监控页面查看任务处理进度

### 文件管理

1. 访问文件管理页面
2. 选择文件并上传
3. 查看文件列表
4. 点击"下载"按钮下载文件
5. 点击"删除"按钮删除文件

### 实时监控

1. 访问实时监控页面
2. 查看系统实时状态（内存、协程数、GC 次数）
3. 查看任务统计（待处理、处理中、已完成）
4. 观察事件日志

## 学习要点

### 1. 数据库操作

**文件**: `internal/database/db.go`, `internal/models/task.go`

```go
// 打开数据库连接
db, err := sql.Open("sqlite", "tasks.db")

// 查询数据
rows, err := db.Query("SELECT * FROM tasks WHERE status = ?", "pending")

// 插入数据
result, err := db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
```

**学习点**:
- `database/sql` 标准库使用
- 预编译语句（? 占位符）
- `defer rows.Close()` 资源释放
- 错误处理

### 2. 文件操作

**文件**: `internal/handlers/file_handler.go`

```go
// 接收上传文件
file, _ := c.FormFile("file")

// 生成唯一文件名
storedName := uuid.New().String() + filepath.Ext(file.Filename)

// 保存文件
c.SaveUploadedFile(file, uploadPath)

// 下载文件
c.File(file.UploadPath)
```

**学习点**:
- `multipart.FileHeader` 处理
- `os.Create`, `os.Remove` 使用
- `filepath.Join` 跨平台路径处理
- `mime.TypeByExtension` MIME 类型识别

### 3. SSE 实时推送

**文件**: `internal/handlers/sse_handler.go`

```go
// 设置 SSE 响应头
c.Header("Content-Type", "text/event-stream")
c.Header("Cache-Control", "no-cache")
c.Header("Connection", "keep-alive")

// 发送事件
c.SSEvent("message", jsonData)
```

**学习点**:
- HTTP 长连接保持
- `text/event-stream` 内容类型
- Goroutine 后台推送数据
- 客户端断开检测

### 4. 协程并发

**文件**: `internal/services/worker_pool.go`

```go
// 创建 Worker Pool
workerPool := NewWorkerPool(5, updateFunc)
workerPool.Start()
defer workerPool.Shutdown()

// 提交任务
workerPool.Submit(job)

// Worker 处理逻辑
func (wp *WorkerPool) worker(id int) {
    for job := range wp.jobs {
        // 处理任务
        processTask(job)
    }
}
```

**学习点**:
- `go` 关键字启动协程
- `chan` 通道创建和使用
- `sync.WaitGroup` 等待协程完成
- `select` 多路复用
- `close(channel)` 关闭通道
- Worker Pool 并发模式

## API 文档

详细的 API 文档请查看 [API.md](docs/API.md)

## 数据库表结构

### tasks 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| title | TEXT | 任务标题 |
| description | TEXT | 任务描述 |
| status | TEXT | 状态（pending, processing, completed, failed） |
| priority | INTEGER | 优先级（1-5） |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |
| completed_at | DATETIME | 完成时间 |

### files 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INTEGER | 主键，自增 |
| original_name | TEXT | 原始文件名 |
| stored_name | TEXT | 存储文件名 |
| file_size | INTEGER | 文件大小（字节） |
| mime_type | TEXT | MIME 类型 |
| upload_path | TEXT | 文件路径 |
| uploaded_at | DATETIME | 上传时间 |

## 常见问题

### 1. 端口 8080 被占用

修改 `cmd/server/main.go` 中的端口号：
```go
srv := &http.Server{
    Addr:    ":8081",  // 改为其他端口
    Handler: router,
}
```

### 2. 数据库文件位置

数据库文件 `tasks.db` 会在运行目录下自动创建。

### 3. 上传文件大小限制

Gin 默认限制为 32MB，可以通过 `router.MaxMultipartMemory` 修改。

### 4. 如何调试

使用 `log.Println()` 或 `fmt.Println()` 输出调试信息。

## 扩展建议

学习完本项目后，可以尝试以下扩展：

1. 添加用户认证（JWT）
2. 添加任务优先级队列
3. 实现任务定时调度
4. 添加数据库迁移工具
5. 添加单元测试
6. 容器化（Docker）
7. 添加配置文件支持
8. 实现 WebSocket 双向通信
9. 添加日志文件管理
10. 实现 API 限流

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题，请通过 GitHub Issues 反馈。

---

**学习愉快！Happy Coding! 🚀**
