package demo

import (
	"fmt"
	"time"

	server "code.byted.org/infcs/mgr/kitex_gen/infcs/mgr/framework"
	"code.byted.org/infcs/mgr/pkg/job"
)

const (
	DeployAction          = "Deploy"
	GetDeployStatusAction = "GetDeployStatus"
)

// Register 返回 Action -> JobStateMachine 的映射，包含 Deploy 与 GetDeployStatus。
func Register() map[string]job.JobStateMachine {
	m := make(map[string]job.JobStateMachine)
	m[DeployAction] = registerDeploy()
	m[GetDeployStatusAction] = registerGetDeployStatus()
	return m
}

// registerDeploy 返回一个 JobStateMachine，包含 PreStage、Stage1、Stage2、Stage3 四个阶段。
// PreStage 同步返回 JobID，Stage1、Stage2、Stage3 异步执行。
func registerDeploy() job.JobStateMachine {
	jm := job.NewJobStateMachine()
	jm.Async = true // 演示异步：PreStage 同步返回，Stage 异步执行
	jm.AddPreStage("PreStage", deployPreStage)
	jm.InitStage = "Stage1"
	jm.AddStage("Stage1", deployStage1)
	jm.AddStage("Stage2", deployStage2)
	jm.AddStage("Stage3", deployStage3)
	return jm
}

// deployPreStage 是一个同步执行的预处理阶段
// 它的存在确保了框架会为异步 Job 生成并返回 JobID
func deployPreStage(jo *job.Job) {
	fmt.Println("[Deploy:PreStage] Running pre-flight checks... JobID will be generated and returned now.")
	// 这里可以放置一些前置检查逻辑，例如参数校验
}

func deployStage1(jo *job.Job) {
	// 演示：读取请求与 App 信息
	fmt.Printf("[Deploy:Stage1] Action=%s\n", jo.Req.MgrReq.Ctx.Action)
	// jo.App 为在 Option.App 注入的业务实例，演示读取信息
	fmt.Printf("[Deploy:Stage1] AppIns=%v\n", jo.App)
	// 模拟工作
	fmt.Println("[Deploy:Stage1] working...")
	time.Sleep(3 * time.Second)
	// 下一阶段
	fmt.Println("[Deploy:Stage1] completed.")
	jo.NextStage = "Stage2"
}

func deployStage2(jo *job.Job) {
	fmt.Printf("[Deploy:Stage2] Action=%s\n", jo.Req.MgrReq.Ctx.Action)
	// 模拟工作
	fmt.Println("[Deploy:Stage2] working...")
	time.Sleep(3 * time.Second)
	// 下一阶段
	fmt.Println("[Deploy:Stage2] completed.")
	jo.NextStage = "Stage3"
}

func deployStage3(jo *job.Job) {
	fmt.Printf("[Deploy:Stage3] Action=%s\n", jo.Req.MgrReq.Ctx.Action)
	// 模拟工作
	fmt.Println("[Deploy:Stage3] working...")
	time.Sleep(3 * time.Second)
	// 标记完成
	jo.SetExit(job.ExitSuccess)
	fmt.Println("[Deploy:Stage3] completed.")
}

// registerGetDeployStatus 返回一个 JobStateMachine，包含 Stage1 一个阶段。
// Stage1 同步执行，根据请求中的 GetJobStatusID 查询 Job 状态并填充响应。
func registerGetDeployStatus() job.JobStateMachine {
	jm := job.NewJobStateMachine()
	jm.Async = false
	jm.AddPreStage("PreStage", getDeployStatusPreStage)
	jm.InitStage = "Stage1"
	jm.AddStage("Stage1", queryJobStatus)
	return jm
}

func getDeployStatusPreStage(jo *job.Job) {
	fmt.Printf("[GetDeployStatus:PreStage] Action=%s\n", jo.Req.MgrReq.Ctx.Action)
	// 这里可以放置一些前置检查逻辑，例如参数校验
}

// 单阶段：根据请求中的 GetJobStatusID 查询 Job 状态并填充响应
func queryJobStatus(jo *job.Job) {
	fmt.Printf("[GetDeployStatus:Stage1] Action=%s\n", jo.Req.MgrReq.Ctx.Action)

	// 模拟工作
	fmt.Println("[GetDeployStatus:Stage1] working...")
	resp := jo.Resp
	resp.MgrResp.Ctx.RequestID = jo.Req.MgrReq.Ctx.RequestID

	// 从 JobManager 获取 Job 实例
	jobID := jo.Req.MgrReq.Ctx.GetJobStatusID
	dst := jo.Jm.GetJob(jobID)
	if dst == nil {
		resp.MgrResp.Ctx.CurStatus = server.JobStatus_Completed
		resp.MgrResp.Ctx.CurStage = ""
		resp.MgrResp.Ctx.GetJobStatusID = jobID
		jo.SetExit(job.ExitSuccess)
		return
	}

	resp.MgrResp.Ctx.CurStatus = dst.GetState4Thrift()
	resp.MgrResp.Ctx.CurStage = dst.GetCurStage()
	resp.MgrResp.Ctx.GetJobStatusID = jobID
	// 如果已经完成，释放内存中的 Job
	if resp.MgrResp.Ctx.CurStatus == server.JobStatus_Completed {
		dst.Free()
	}
	// 标记完成
	jo.SetExit(job.ExitSuccess)
	fmt.Println("[GetDeployStatus:Stage1] completed.")
}
