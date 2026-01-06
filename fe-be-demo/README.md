# Go 后端调用 Demo

本项目是一个基于 Go 标准库实现的简单 HTTP 后端服务，旨在演示常见的前后端 API 调用模式。

## 目录结构

```
.
├── backend/
│   ├── main.go         # 后端主程序
│   ├── go.mod
│   └── README.md       # 本文档
└── frontend/
    └── index.html      # 前端交互页面
```

## 功能特性

- **轻量级实现**：仅使用 Go 标准库（`net/http`），无第三方框架依赖。
- **RESTful API**：提供符合 REST 风格的接口。
- **统一响应格式**：所有接口均返回统一的 JSON 结构 `{ code, message, data }`。
- **CORS 支持**：内置中间件，支持来自 `http://localhost:5173` 和本地文件 (`file://`) 的跨域请求。
- **内存存储**：`todos` 接口的数据存储在内存中，服务重启后会重置。
- **简单日志**：记录每个请求的方法、路径和处理耗时。

## 如何运行

### 环境要求

- Go 1.20 或更高版本。

### 启动后端服务

1.  打开终端，进入 `backend` 目录。

    ```bash
    cd backend
    ```

2.  运行 `main.go` 文件。

    ```bash
    go run main.go
    ```

3.  服务成功启动后，你将看到以下日志输出：

    ```
    YYYY/MM/DD hh:mm:ss HTTP server starting on :8080...
    ```

    此时，后端服务正在本地 `8080` 端口监听请求。

### 访问前端页面

1.  在浏览器中直接打开 `frontend/index.html` 文件。
2.  页面中的“后端 Base URL”默认已设置为 `http://localhost:8080`。
3.  点击页面上的按钮，即可与本地运行的后端服务进行交互，并在“返回结果”区域查看完整的 HTTP 响应详情。

## 接口规范

- **Base URL**: `http://localhost:8080`
- **路径前缀**: `/api`

### 统一返回格式

所有接口在成功或失败时都返回如下格式的 JSON 对象：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "key": "value"
  }
}
```

- `code`: 状态码。`0` 表示成功，非 `0` 表示失败。
- `message`: 描述信息。成功时通常为 `"ok"`，失败时为具体的错误说明。
- `data`: 实际的业务数据。如果接口没有数据返回，此字段可能被省略。

### Endpoints

#### 1. 健康检查

- **Endpoint**: `GET /api/health`
- **描述**: 检查服务是否正常运行。
- **成功响应** (`200 OK`):
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "status": "ok"
    }
  }
  ```

#### 2. Echo 服务

- **Endpoint**: `POST /api/echo`
- **描述**: 原样返回请求体中的 JSON 数据，并附加一个 `server_time` 字段。
- **请求体** (JSON):
  ```json
  {
    "message": "hello from frontend"
  }
  ```
- **成功响应** (`200 OK`):
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "message": "hello from frontend",
      "server_time": "2023-10-27T10:00:00Z"
    }
  }
  ```
- **失败响应** (`400 Bad Request`，若请求体非 JSON 格式):
  ```json
  {
    "code": 400,
    "message": "invalid JSON body",
    "data": {
      "detail": "..."
    }
  }
  ```

#### 3. 获取 Todo 列表

- **Endpoint**: `GET /api/todos`
- **描述**: 返回当前存储在内存中的所有 `todo` 事项。
- **成功响应** (`200 OK`):
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": [
      {
        "id": 1,
        "title": "学习 Go",
        "done": false
      }
    ]
  }
  ```

#### 4. 新增 Todo

- **Endpoint**: `POST /api/todos`
- **描述**: 创建一个新的 `todo` 事项。ID 会在服务端自动递增。
- **请求体** (JSON): `title` 为必填字段。
  ```json
  {
    "title": "编写 README"
  }
  ```
- **成功响应** (`200 OK`): 返回新创建的 `todo` 对象。
  ```json

  {
    "code": 0,
    "message": "ok",
    "data": {
      "id": 2,
      "title": "编写 README",
      "done": false
    }
  }
  ```
- **失败响应** (`400 Bad Request`，若 `title` 缺失):
  ```json
  {
    "code": 400,
    "message": "title is required"
  }
  ```

## 注意事项

- **CORS 策略**:
    - 后端允许来自 `http://localhost:5173` 的跨域请求，这是常见的前端开发服务器地址。
    - 同时，也允许 `Origin: null` 的请求，这通常对应于直接在浏览器中打开的本地 `file://` 文件。
    - 其他来源的跨域请求默认会被拒绝。

- **数据持久化**:
    - 本 Demo 的 `todos` 列表存储在内存中。这意味着每次重启后端服务，所有 `todo` 数据都会丢失。这是一种简化的做法，仅用于演示 API 功能。

- **错误处理**:
    - Demo 中包含了基础的错误处理，如请求方法不被允许 (`405 Method Not Allowed`)、JSON 解析失败 (`400 Bad Request`) 等，并通过统一的 JSON 格式返回错误信息。
