package demo

import (
	"fmt"
	"time"

	server "code.byted.org/infcs/mgr/kitex_gen/infcs/mgr/framework"
	"code.byted.org/infcs/mgr/pkg/job"
)

// Register 返回 Action -> JobStateMachine 的映射，包含 Deploy 与 GetDeployStatus。
func Register() map[string]job.JobStateMachine {
	m := make(map[string]job.JobStateMachine)
	m["Deploy"] = registerDeploy()
	m["GetDeployStatus"] = registerGetDeployStatus()
	return m
}

// --- Deploy ---
func registerDeploy() job.JobStateMachine {
	jm := job.NewJobStateMachine()
	jm.Async = true // 演示异步：PreStage 同步返回，Stage 异步执行
	jm.InitStage = "Stage1"
	jm.AddStage("Stage1", deployStage1)
	jm.AddStage("Stage2", deployStage2)
	return jm
}

func deployStage1(jo *job.Job) {
	// 演示：读取请求与 App 信息
	fmt.Printf("[Deploy:Stage1] Action=%s\n", jo.Req.MgrReq.Ctx.Action)
	// jo.App 为在 Option.App 注入的业务实例，演示读取信息
	fmt.Printf("[Deploy:Stage1] AppIns=%v\n", jo.App)
	// 模拟工作
	fmt.Println("[Deploy:Stage1] working...")
	time.Sleep(1 * time.Second)
	// 下一阶段
	jo.NextStage = "Stage2"
}

func deployStage2(jo *job.Job) {
	fmt.Printf("[Deploy:Stage2] Action=%s\n", jo.Req.MgrReq.Ctx.Action)
	fmt.Println("[Deploy:Stage2] working...")
	time.Sleep(1 * time.Second)
	// 标记完成
	jo.SetExit(job.ExitSuccess)
}

// --- GetDeployStatus ---
func registerGetDeployStatus() job.JobStateMachine {
	jm := job.NewJobStateMachine()
	jm.Async = false
	jm.InitStage = "Stage1"
	jm.AddStage("Stage1", queryJobStatus)
	return jm
}

// 单阶段：根据请求中的 GetJobStatusID 查询 Job 状态并填充响应
func queryJobStatus(jo *job.Job) {
	resp := jo.Resp
	resp.MgrResp.Ctx.RequestID = jo.Req.MgrReq.Ctx.RequestID

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
	jo.SetExit(job.ExitSuccess)
}
