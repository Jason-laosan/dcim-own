# TaskHub API 文档

本文档详细说明了 TaskHub 项目的所有 API 端点。

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **Content-Type**: `application/json`
- **编码**: UTF-8

---

## 任务管理 API

### 1. 获取所有任务

获取任务列表，支持按状态过滤。

**端点**: `GET /api/tasks`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| status | string | 否 | 过滤状态（pending, processing, completed, failed） |

**请求示例**:
```bash
# 获取所有任务
curl http://localhost:8080/api/tasks

# 获取待处理任务
curl http://localhost:8080/api/tasks?status=pending
```

**响应示例**:
```json
[
  {
    "id": 1,
    "title": "学习 Go 语言",
    "description": "完成 Go 基础教程",
    "status": "pending",
    "priority": 3,
    "created_at": "2024-01-06T10:30:00Z",
    "updated_at": "2024-01-06T10:30:00Z",
    "completed_at": null
  },
  {
    "id": 2,
    "title": "实现文件上传功能",
    "description": "使用 multipart/form-data",
    "status": "completed",
    "priority": 5,
    "created_at": "2024-01-06T09:00:00Z",
    "updated_at": "2024-01-06T11:00:00Z",
    "completed_at": "2024-01-06T11:00:00Z"
  }
]
```

---

### 2. 获取单个任务

根据 ID 获取特定任务的详细信息。

**端点**: `GET /api/tasks/:id`

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 任务 ID |

**请求示例**:
```bash
curl http://localhost:8080/api/tasks/1
```

**响应示例**:
```json
{
  "id": 1,
  "title": "学习 Go 语言",
  "description": "完成 Go 基础教程",
  "status": "pending",
  "priority": 3,
  "created_at": "2024-01-06T10:30:00Z",
  "updated_at": "2024-01-06T10:30:00Z",
  "completed_at": null
}
```

**错误响应**:
```json
{
  "error": "Task not found"
}
```

---

### 3. 创建任务

创建新的任务。

**端点**: `POST /api/tasks`

**请求体**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 任务标题 |
| description | string | 否 | 任务描述 |
| priority | integer | 否 | 优先级（1-5，默认 1） |

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "实现 SSE 推送",
    "description": "使用 Server-Sent Events",
    "priority": 4
  }'
```

**响应示例**:
```json
{
  "id": 3,
  "title": "实现 SSE 推送",
  "description": "使用 Server-Sent Events",
  "status": "pending",
  "priority": 4,
  "created_at": "2024-01-06T12:00:00Z",
  "updated_at": "2024-01-06T12:00:00Z",
  "completed_at": null
}
```

---

### 4. 更新任务

更新现有任务的信息。

**端点**: `PUT /api/tasks/:id`

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 任务 ID |

**请求体**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| title | string | 是 | 任务标题 |
| description | string | 否 | 任务描述 |
| status | string | 否 | 状态 |
| priority | integer | 否 | 优先级 |

**请求示例**:
```bash
curl -X PUT http://localhost:8080/api/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "学习 Go 语言（已更新）",
    "description": "完成 Go 高级教程",
    "status": "processing",
    "priority": 5
  }'
```

**响应示例**:
```json
{
  "id": 1,
  "title": "学习 Go 语言（已更新）",
  "description": "完成 Go 高级教程",
  "status": "processing",
  "priority": 5,
  "created_at": "2024-01-06T10:30:00Z",
  "updated_at": "2024-01-06T13:00:00Z",
  "completed_at": null
}
```

---

### 5. 删除任务

删除指定的任务。

**端点**: `DELETE /api/tasks/:id`

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 任务 ID |

**请求示例**:
```bash
curl -X DELETE http://localhost:8080/api/tasks/1
```

**响应示例**:
```json
{
  "message": "Task deleted successfully"
}
```

---

### 6. 异步处理任务

将任务提交到 Worker Pool 进行异步处理（演示协程并发）。

**端点**: `POST /api/tasks/:id/process`

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 任务 ID |

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/tasks/1/process
```

**响应示例**:
```json
{
  "message": "任务已提交到 Worker Pool 进行异步处理",
  "task_id": 1
}
```

**说明**:
- 任务会被提交到后台 Worker Pool
- 状态会自动更新为 `processing`
- 处理完成后状态更新为 `completed` 或 `failed`
- 可在实时监控页面查看处理进度

---

## 文件管理 API

### 1. 获取文件列表

获取所有已上传文件的列表。

**端点**: `GET /api/files`

**请求示例**:
```bash
curl http://localhost:8080/api/files
```

**响应示例**:
```json
[
  {
    "id": 1,
    "original_name": "document.pdf",
    "stored_name": "a1b2c3d4-e5f6-4789-a0b1-c2d3e4f5g6h7.pdf",
    "file_size": 102400,
    "mime_type": "application/pdf",
    "upload_path": "uploads/a1b2c3d4-e5f6-4789-a0b1-c2d3e4f5g6h7.pdf",
    "uploaded_at": "2024-01-06T14:00:00Z"
  },
  {
    "id": 2,
    "original_name": "image.png",
    "stored_name": "b2c3d4e5-f6g7-4890-h1i2-j3k4l5m6n7o8.png",
    "file_size": 51200,
    "mime_type": "image/png",
    "upload_path": "uploads/b2c3d4e5-f6g7-4890-h1i2-j3k4l5m6n7o8.png",
    "uploaded_at": "2024-01-06T14:30:00Z"
  }
]
```

---

### 2. 上传文件

上传文件到服务器。

**端点**: `POST /api/files/upload`

**Content-Type**: `multipart/form-data`

**表单字段**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | file | 是 | 要上传的文件 |

**请求示例**:
```bash
curl -X POST http://localhost:8080/api/files/upload \
  -F "file=@/path/to/your/file.pdf"
```

**响应示例**:
```json
{
  "message": "File uploaded successfully",
  "file": {
    "id": 3,
    "original_name": "file.pdf",
    "stored_name": "c3d4e5f6-g7h8-4901-i2j3-k4l5m6n7o8p9.pdf",
    "file_size": 204800,
    "mime_type": "application/pdf",
    "upload_path": "uploads/c3d4e5f6-g7h8-4901-i2j3-k4l5m6n7o8p9.pdf",
    "uploaded_at": "2024-01-06T15:00:00Z"
  }
}
```

**错误响应**:
```json
{
  "error": "No file uploaded",
  "details": "multipart: no file"
}
```

---

### 3. 下载文件

下载指定的文件。

**端点**: `GET /api/files/:id/download`

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 文件 ID |

**请求示例**:
```bash
# 使用浏览器访问
http://localhost:8080/api/files/1/download

# 使用 curl 下载
curl -O -J http://localhost:8080/api/files/1/download
```

**响应**:
- 成功：返回文件流
- 失败：返回 JSON 错误信息

---

### 4. 删除文件

删除指定的文件（同时删除数据库记录和磁盘文件）。

**端点**: `DELETE /api/files/:id`

**路径参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| id | integer | 文件 ID |

**请求示例**:
```bash
curl -X DELETE http://localhost:8080/api/files/1
```

**响应示例**:
```json
{
  "message": "File deleted successfully"
}
```

---

## SSE 实时推送 API

### 1. 任务状态更新推送

订阅任务状态变化的实时推送。

**端点**: `GET /api/sse/tasks`

**协议**: Server-Sent Events

**事件格式**:
```
data: {"timestamp":"2024-01-06T16:00:00Z","pending_count":5,"processing_count":2,"completed_count":10}
```

**JavaScript 客户端示例**:
```javascript
const eventSource = new EventSource('/api/sse/tasks');

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('任务状态:', data);

  // data.pending_count - 待处理任务数
  // data.processing_count - 处理中任务数
  // data.completed_count - 已完成任务数
};

eventSource.onerror = function(error) {
  console.error('连接错误:', error);
};
```

**推送频率**: 每 2 秒

---

### 2. 系统状态推送

订阅系统运行状态的实时推送。

**端点**: `GET /api/sse/system`

**协议**: Server-Sent Events

**事件格式**:
```
data: {"timestamp":"2024-01-06T16:00:00Z","goroutines":15,"memory_mb":12,"memory_total_mb":150,"gc_count":5,"processing_tasks":2}
```

**JavaScript 客户端示例**:
```javascript
const eventSource = new EventSource('/api/sse/system');

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('系统状态:', data);

  // data.goroutines - 当前协程数量
  // data.memory_mb - 当前内存使用（MB）
  // data.memory_total_mb - 累计分配内存（MB）
  // data.gc_count - GC 执行次数
  // data.processing_tasks - 处理中任务数
};

eventSource.onerror = function(error) {
  console.error('连接错误:', error);
  // 可以实现自动重连
};
```

**推送频率**: 每 1 秒

---

## Web 页面路由

### 1. 主页
- **路径**: `GET /`
- **说明**: 项目介绍和功能展示

### 2. 任务管理页面
- **路径**: `GET /tasks`
- **说明**: 任务的创建、查看、处理、删除

### 3. 文件管理页面
- **路径**: `GET /files`
- **说明**: 文件的上传、下载、删除

### 4. 实时监控页面
- **路径**: `GET /monitor`
- **说明**: 系统状态和任务状态的实时监控

---

## 错误码

| HTTP 状态码 | 说明 |
|------------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 202 | 已接受（异步任务） |
| 400 | 请求参数错误 |
| 404 | 资源不找到 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 |

---

## 使用示例

### 完整工作流示例

```bash
# 1. 创建任务
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "测试任务", "priority": 3}'

# 2. 获取任务列表
curl http://localhost:8080/api/tasks

# 3. 提交任务到 Worker Pool 异步处理
curl -X POST http://localhost:8080/api/tasks/1/process

# 4. 上传文件
curl -X POST http://localhost:8080/api/files/upload \
  -F "file=@test.pdf"

# 5. 获取文件列表
curl http://localhost:8080/api/files

# 6. 下载文件
curl -O -J http://localhost:8080/api/files/1/download

# 7. 删除文件
curl -X DELETE http://localhost:8080/api/files/1

# 8. 删除任务
curl -X DELETE http://localhost:8080/api/tasks/1
```

---

## 注意事项

1. **并发处理**: 任务提交到 Worker Pool 后是异步处理的，状态变化可通过 SSE 实时监控
2. **文件大小**: 默认限制 32MB，可通过配置修改
3. **CORS**: 已配置允许跨域访问
4. **日志**: 所有请求都会记录在服务器日志中

---

**更多问题请查看 [README.md](README.md)**
