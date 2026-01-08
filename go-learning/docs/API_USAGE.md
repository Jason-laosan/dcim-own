# API 使用文档

## Swagger 文档

项目已集成 Swagger API 文档，启动服务器后访问：

```
http://localhost:8080/swagger/index.html
```

在 Swagger UI 中可以查看所有 API 端点、请求参数和响应格式，并直接测试 API。

## 认证功能

### 1. 用户注册

**端点**: `POST /api/auth/register`

**请求体**:
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "created_at": "2026-01-08T10:00:00Z",
    "updated_at": "2026-01-08T10:00:00Z"
  }
}
```

**使用 curl**:
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### 2. 用户登录

**端点**: `POST /api/auth/login`

**请求体**:
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "created_at": "2026-01-08T10:00:00Z",
    "updated_at": "2026-01-08T10:00:00Z"
  }
}
```

**使用 curl**:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### 3. 获取用户信息

**端点**: `GET /api/auth/profile`

**需要认证**: 是

**请求头**:
```
Authorization: Bearer <your-jwt-token>
```

**响应**:
```json
{
  "id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "created_at": "2026-01-08T10:00:00Z",
  "updated_at": "2026-01-08T10:00:00Z"
}
```

**使用 curl**:
```bash
curl -X GET http://localhost:8080/api/auth/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## JWT Token 使用

### Token 格式

所有需要认证的 API 请求都需要在请求头中包含 JWT token：

```
Authorization: Bearer <your-jwt-token>
```

### Token 有效期

- JWT token 有效期为 24 小时
- Token 过期后需要重新登录获取新的 token

### 在 Swagger UI 中使用认证

1. 访问 Swagger UI: `http://localhost:8080/swagger/index.html`
2. 点击右上角的 "Authorize" 按钮
3. 在弹出的对话框中输入: `Bearer <your-jwt-token>`
4. 点击 "Authorize" 按钮
5. 现在可以测试需要认证的 API 端点

## 安全注意事项

⚠️ **重要**: 当前使用的 JWT 密钥是硬编码的示例密钥。在生产环境中，请务必：

1. 将 JWT 密钥改为强随机字符串
2. 使用环境变量存储密钥
3. 定期更换密钥
4. 使用 HTTPS 加密传输

修改 JWT 密钥位置：`internal/middleware/auth.go`

```go
var jwtSecret = []byte("your-secret-key-change-this-in-production")
```

建议改为从环境变量读取：

```go
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
```

## 数据库

用户数据存储在 SQLite 数据库的 `users` 表中：

- 用户名必须唯一
- 邮箱必须唯一
- 密码使用 bcrypt 加密存储

## 测试流程

### 完整测试流程示例

```bash
# 1. 注册新用户
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "test123456"
  }'

# 保存返回的 token

# 2. 使用 token 获取用户信息
curl -X GET http://localhost:8080/api/auth/profile \
  -H "Authorization: Bearer <your-token-here>"

# 3. 登录（使用相同的邮箱和密码）
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "test123456"
  }'
```

## 错误处理

### 常见错误响应

**400 Bad Request** - 请求参数错误
```json
{
  "error": "validation error message"
}
```

**401 Unauthorized** - 未认证或 token 无效
```json
{
  "error": "Invalid or expired token"
}
```

**409 Conflict** - 用户名或邮箱已存在
```json
{
  "error": "Email already registered"
}
```

**500 Internal Server Error** - 服务器错误
```json
{
  "error": "Internal server error"
}
```
