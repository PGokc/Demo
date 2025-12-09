# MGR 框架入门 Demo

本项目是一个简化的演示，旨在阐明像 `code.byted.org/infcs/mgr` 这类服务管理框架的基础概念。

## 核心概念演示

1.  **应用生命周期管理**: `framework/mgr.go` 文件模拟了核心应用，它负责管理所有已注册组件的启动和优雅关闭。

2.  **组件化架构**: 整个应用由模块化的 `Component` 构建而成。`components/http_server.go` 是一个具体的例子。每个组件都包含 `Init`, `Start`, `Stop` 等方法。

3.  **集中式注册**: 在 `main.go` 中，所有组件被注册到中央应用实例中，由它来统一控制。

## 项目结构

- `main.go`: 应用的入口。它负责初始化 `mgr` 应用并注册所有组件。
- `go.mod`: Go 模块定义文件。
- `README.md`: 本说明文档。
- `framework/`: 此目录包含我们模拟的 `mgr` 框架。
    - `mgr.go`: 模拟了核心的应用运行器。
- `components/`: 此目录包含具体的业务逻辑组件。
    - `http_server.go`: 一个运行简单 Gin Web 服务的示例组件。

## 如何运行

1.  在终端中进入此项目目录：
    ```bash
    cd /Users/bytedance/Lark-Demo/mgr-demo
    ```

2.  整理并下载依赖：
    ```bash
    go mod tidy
    ```

3.  运行应用：
    ```bash
    go run main.go
    ```

4.  您将看到日志输出，提示 HTTP 服务器已在 8080 端口启动。

5.  打开一个新的终端，测试 API 端点：
    ```bash
    curl http://localhost:8080/hello
    ```
    您应该会收到响应: `{"message":"Hello from the MGR-powered component!"}`
