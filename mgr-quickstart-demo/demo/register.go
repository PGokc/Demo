package demo

import (
	"fmt"
	"time"

	"mgr-quickstart-demo/app"

	"code.byted.org/infcs/mgr/pkg/job"
)

const (
	DeployAction          = "Deploy"
	GetDeployStatusAction = "GetDeployStatus"
)

// RegisterJobFunc 将业务动作注册到 Mgr，示例包含两个 Action：Deploy 与 GetDeployStatus
func RegisterJobFunc() (Register map[string]job.JobStateMachine) {
	Register = make(map[string]job.JobStateMachine)

	// 注册部署与查询部署状态的 JobStateMachine
	Register[DeployAction] = DeployActionJobStateMachine()
	Register[GetDeployStatusAction] = DeployStatusJobStateMachine()

	return Register
}

// DeployActionJobStateMachine 示例：包含两个 Stage：Stage1、Stage2
func DeployActionJobStateMachine() job.JobStateMachine {
	preStageMap := make(map[string]job.StageFunc)
	preStageMap["PreStage1"] = PreDeployStage1

	stageMap := make(map[string]job.StageFunc)
	stageMap["Stage1"] = DeployStage1
	stageMap["Stage2"] = DeployStage2

	jobStateMachine := job.JobStateMachine{
		PreStage:  preStageMap,
		InitStage: "Stage1",
		Stage:     stageMap,
		Async:     true,
	}
	return jobStateMachine
}

// PreDeployStage1 示例：打印日志、演示私有实例使用，并设置下一阶段
func PreDeployStage1(j *job.Job) {
	fmt.Println("Start to run PreDeploy Stage1")
	fmt.Printf("Got req, Action: %s\n", j.Req.MgrReq.Ctx.Action)
}

// DeployStage1 示例：打印日志、演示私有实例使用，并设置下一阶段
func DeployStage1(j *job.Job) {
	req := j.Req
	fmt.Println("Start to run Deploy Stage1")
	fmt.Printf("Got req, Action: %s\n", req.MgrReq.Ctx.Action)

	// 通过 job.App 获取私有实例（用户自定义），并调用其方法
	if j.App != nil {
		if appIns, ok := j.App.(app.App); ok {
			appIns.Install(j)
		}
	}

	// 模拟 Stage1 耗时 6s
	for count := 0; count < 2; count++ {
		fmt.Println("Just working on Stage1.")
		time.Sleep(3 * time.Second)
	}
	fmt.Println("Deploy Stage1 done.")

	// 设置下一阶段为 Stage2
	j.NextStage = "Stage2"
}

// DeployStage2 示例：打印日志并结束 Job
func DeployStage2(j *job.Job) {
	req := j.Req
	fmt.Println("Start to run Deploy Stage2")
	fmt.Printf("Got req, Action: %s\n", req.MgrReq.Ctx.Action)

	// 模拟 Stage2 耗时 6s
	for count := 0; count < 2; count++ {
		fmt.Println("Just working on Stage2.")
		time.Sleep(3 * time.Second)
	}
	fmt.Println("Deploy Stage2 done.")

	// 设置该Job状态为 Completed
	j.SetExit(job.ExitSuccess)
}

// DeployStatusJobStateMachine 示例：包含一个 Stage：Stage1
func DeployStatusJobStateMachine() job.JobStateMachine {
	preStageMap := make(map[string]job.StageFunc)
	preStageMap["PreStage1"] = PreGetDeployStatusStage

	stageMap := make(map[string]job.StageFunc)
	stageMap["Stage1"] = GetDeployStatusStage

	jobStateMachine := job.JobStateMachine{
		PreStage:  preStageMap,
		InitStage: "Stage1",
		Stage:     stageMap,
	}
	return jobStateMachine
}

// PreGetDeployStatusStage 示例：打印日志、演示私有实例使用，并设置下一阶段
func PreGetDeployStatusStage(j *job.Job) {
	fmt.Println("Start to run PreGetDeployStatus Stage1")
	fmt.Printf("Got req, Action: %s\n", j.Req.MgrReq.Ctx.Action)
}

// GetDeployStatusStage 示例：查询 Job 状态并返回，同时标记Job为 Completed
func GetDeployStatusStage(j *job.Job) {
	fmt.Println("Get into GetDeployStatus.")

	// 通过jobManager查询Job状态
	fmt.Printf("Query Job %s Status.\n", j.Req.MgrReq.Ctx.GetJobStatusID)
	dstJob := j.Jm.GetJob(j.Req.MgrReq.Ctx.GetJobStatusID)

	// 打印Job结构体信息
	fmt.Printf("Job %s Info: %+v.\n", dstJob.GetJobIdStr(), dstJob)

	// 如果Job状态为 Completed，则释放Job资源
	if dstJob.GetStateStr() == "Completed" {
		dstJob.Free()
	}

	// 设置当前Job状态为 Completed
	j.SetExit(job.ExitSuccess)
}
