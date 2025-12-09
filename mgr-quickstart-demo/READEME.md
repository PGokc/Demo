# Mgr 快速入门示例

该示例工程演示如何在服务端集成并启动 Mgr 框架、注册两阶段的业务 Job、演示私有实例注入，以及提供 Kitex 客户端调用参考。

- 示例目录：`mgr-quickstart-demo/`
- 监听端口：`:8888`
- 依赖库：`code.byted.org/infcs/mgr/pkg/mgr`、`pkg/utils`、`pkg/job`
- 代码可编译运行的前提：在字节内网配置好依赖（GOPRIVATE、拉取 `infcs/mgr` 最新 tag、Kitex 生成代码与 yml）

## 环境准备

1. 配置私有模块拉取：

```bash
go env -w GOPRIVATE=code.byted.org
```

2. 在内网拉取 Mgr 依赖（使用最新版 tag）：

```bash
# 请替换为当前可用的最新 tag
go get code.byted.org/infcs/mgr@latest
```

3. 如需运行客户端示例，准备好 Kitex IDL 并生成代码与 `kitex.yml`。

## 启动服务端

单副本（无选主）：

```bash
go run main.go --id s1 --electionFlag 0 --groupName MgrQuickstart --groupId 123
```

三副本选主（需配置 ZK 集群）：

```bash
# 三个终端分别执行：
go run main.go --id s1 --electionFlag 1 --zkAdress 10.227.31.8:2181 --groupName MgrQuickstart --groupId 123
go run main.go --id s2 --electionFlag 1 --zkAdress 10.227.31.8:2181 --groupName MgrQuickstart --groupId 123
go run main.go --id s3 --electionFlag 1 --zkAdress 10.227.31.8:2181 --groupName MgrQuickstart --groupId 123
```

可选参数：

```bash
--sessionTimeout 8  # 单位秒，默认 6s
```

## 注册的业务 Job

- `Deploy` 两阶段：`Stage1` 日志 + 私有实例调用 + `NextStage="Stage2"`；`Stage2` 日志 + `Finished=true`
- `GetDeployStatus` 单阶段：查询状态并返回，`Finished=true`

## 客户端调用示例（Kitex）

- 参考 `client/mock_client.go`，该文件使用 `//go:build example` 构建标签，默认不会参与构建。
- 在内网准备好 Kitex 生成的客户端代码与 `kitex.yml` 后，将示例中的 Request 与 Client API 替换为你生成的版本，并使用：

```go
resp, err := cli.Action(ctx, req, callopt.WithHostPort("127.0.0.1:8888"))
```

## 异步调用与长连接

- 异步：`MgrReq.Async = true`；服务端快速返回 ack，客户端随后使用 `JobID` 轮询 `GetDeployStatus`。
- 长连接：对于耗时较长的同步请求，优先选择长连接以避免 RPC 连接超时。

## 运行期输出预期

- 服务端启动后日志包含监听地址与选主信息；
- 执行 `Deploy Stage1/Stage2` 时打印阶段日志；
- 异步模式下客户端快速收到 ack；轮询时看到 `JobStatus` 与 `JobStage` 的变化。

## 免责声明

- 本示例依赖字节内网模块与 Kitex 生成代码；请在内网准备好依赖后再进行编译运行。
