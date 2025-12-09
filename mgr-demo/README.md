# Mgr 快速入门 Demo

本示例构建了一个最小可运行工程，演示如何基于字节内部的 `mgr` 框架注册 Action、编排阶段并启动服务，同时mock了一个客户端调用，方便初学者快速跑通。

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

## 目录结构

- `main.go`：解析 flag，装配 `mgr.Option`，启动服务。
- `demo/register.go`：注册 `Deploy` 与 `GetDeployStatus` 的状态机，演示阶段流转与状态查询。
- `app/app.go`：定义一个轻量的 `App` 作为 `job.AppIns` 注入，供阶段内读取。
- `client/client.go`：Kitex 客户端调用示例骨架（构建标签 `example`）。

