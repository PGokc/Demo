# Mgr 快速入门 Demo v2

本示例在本地 Workspace 构建了一个最小可运行工程，演示如何基于字节内网 `mgr` 框架注册 Action、编排阶段并启动服务，提供 Kitex 客户端调用骨架，方便初学者快速跑通。

## 环境准备

- Go 环境：推荐 Go 1.20+。
- 配置 GOPRIVATE（字节内网）：
    - 在终端执行：
        - `go env -w GOPRIVATE=code.byted.org`
- 拉取依赖 `infcs/mgr` 最新 tag：
    - 在内网环境执行：
        - `go get code.byted.org/infcs/mgr@latest`
    - 本工程的 `go.mod` 已使用 `replace code.byted.org/infcs/mgr => mgr` 指向本地已下载的仓库，便于在当前 Workspace 编译；在内网环境请删除 replace 并使用上面的 go get。
- Kitex 客户端生成说明：
    - 需要准备框架 IDL 与 `kitex.yml`，例如 `infcs/mgr/framework.thrift`。
    - 在工程根目录执行：`kitex -module mgr-quickstart-demo-v2 -service mgr-client infcs/mgr/framework.thrift`。
    - 根据生成的客户端调用服务，参考 `client/mock_client.go`。

## 启动示例

- 单副本（不选主）：
    - `go run ./main.go --id=demo-1 --electionFlag=false --groupName=MgrQuickStart`
    - 端口：默认使用 `:8889`（`utils.DefaultAddress`）。
- 三副本选主（ZK 选主示例）：
    - 分别在三台或本地三个进程运行（ID 不同）：
        - `go run ./main.go --id=demo-1 --electionFlag=true --zkAddress=zk1:2181,zk2:2181 --groupName=MgrQuickStart`
        - `go run ./main.go --id=demo-2 --electionFlag=true --zkAddress=zk1:2181,zk2:2181 --groupName=MgrQuickStart`
        - `go run ./main.go --id=demo-3 --electionFlag=true --zkAddress=zk1:2181,zk2:2181 --groupName=MgrQuickStart`
    - 或使用多次传参：`--zkAddress=zk1:2181 --zkAddress=zk2:2181`

## Action 注册与阶段流转

- 在 `demo/register.go` 中通过 `Register()` 返回 `map[string]job.JobStateMachine`，注册两个 Action：
    - `Deploy`：两阶段
        - `Stage1`：日志输出，演示读取 `job.Req.MgrReq.Ctx.Action` 与 `job.App` 信息，设置 `job.NextStage = "Stage2"`。
        - `Stage2`：日志输出，调用 `job.SetExit(job.ExitSuccess)` 标记完成。
    - `GetDeployStatus`：单阶段
        - 读取 `job.Req.MgrReq.Ctx.GetJobStatusID`，使用 `jo.Jm.GetJob(jobID)` 查询本地队列中的 Job；不存在则视为完成。
        - 将结果填充到 `job.Resp.MgrResp.Ctx`：`CurStatus`、`CurStage`、`GetJobStatusID`；若已完成则 `dst.Free()` 释放。
- 关键要点：
    - 阶段跳转使用 `job.NextStage`；结束使用 `job.SetExit(job.ExitSuccess)`。
    - `JobStateMachine.Async=true` 表示异步模式：`PreStage` 同步返回，`Stage` 异步执行。

## 同步与异步模式

- 同步模式：`JobStateMachine.Async=false`，`Stage` 执行完才返回。
- 异步模式：`JobStateMachine.Async=true`，`Stage` 异步执行，客户端可轮询 `GetDeployStatus`。
- 轮询示例：
    - 首次调用 `Deploy` 获取 `RequestID` 作为 JobID；随后构造请求：`req.MgrReq.Ctx.Action = "GetDeployStatus"`，`req.MgrReq.Ctx.GetJobStatusID = <JobID>`，循环调用 `cli.Action(ctx, req, callopt.WithHostPort("127.0.0.1:8889"))` 直到 `resp.MgrResp.Ctx.CurStatus == Completed`。

## 客户端调用骨架

- 文件 `client/mock_client.go` 使用 `//go:build example` 构建标签，提供 Kitex 调用的示例骨架。
    - 说明需要在内网准备 IDL 与 `kitex.yml` 并生成客户端代码。
    - 给出 `cli.Action(ctx, req, callopt.WithHostPort("127.0.0.1:8889"))` 的调用示例。

## 运行期输出预期与常见问题

- 运行期输出：
    - 启动时打印 Mgr 参数、Action 与 Stage 日志，例如 `[Deploy:Stage1] working...`、`[Deploy:Stage2] working...`。
    - `GetDeployStatus` 返回 `CurStatus` 与 `CurStage`，完成后可能释放本地 Job。
- 常见问题：
    - RPC 超时：对于耗时长的任务，建议采用异步模式并通过轮询 `GetDeployStatus` 查询状态；或使用 Kitex 长连接与较大的超时时间（`WithLongConnection`、`WithRPCTimeout`）。
    - 端口占用：默认监听 `:8889`，如需调整请修改 `main.go` 中的 `Address` 设置为 `utils.NewNetAddr("tcp", ":<port>")`。

## 目录结构

- `main.go`：解析 flag，装配 `mgr.Option`，启动服务。
- `demo/register.go`：注册 `Deploy` 与 `GetDeployStatus` 的状态机，演示阶段流转与状态查询。
- `app/app.go`：定义一个轻量的 `App` 作为 `job.AppIns` 注入，供阶段内读取。
- `client/mock_client.go`：Kitex 客户端调用示例骨架（构建标签 `example`）。
- `go.mod`：模块名与依赖；在 Workspace 内通过 `replace` 指向本地 `mgr` 仓库以便编译。

## 注意

- 导入路径务必使用真实路径：`code.byted.org/infcs/mgr/pkg/mgr`、`code.byted.org/infcs/mgr/pkg/utils`、`code.byted.org/infcs/mgr/pkg/job` 等。
- 示例以字节内网依赖为前提，编译与运行取决于内网环境与 IDL 生成是否就绪。
