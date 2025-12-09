package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	log "code.byted.org/infcs/lib-log"
	"code.byted.org/infcs/mgr/pkg/job"
	"code.byted.org/infcs/mgr/pkg/mgr"
	"code.byted.org/infcs/mgr/pkg/utils"
)

// 客户端请求的数据结构
type PrintRequest struct {
	Message string `json:"message"`
}

// 在阶段间传递数据的上下文结构体
type JobContext struct {
	RequestData PrintRequest
	ProcessTime time.Time
}

// --- 作业阶段处理函数 ---

// 阶段1: 解析请求
func stageParseRequest(jo *job.Job) {
	fmt.Println("====== [Stage 1/3: ParseRequest] ======")

	reqBody := jo.Req.PrivateReq.ReqBytes
	if len(reqBody) == 0 {
		fmt.Println("Error: Request body (PrivateReq) is empty")
		jo.SetExit(job.ExitErr)
		return
	}

	var printReq PrintRequest
	if err := json.Unmarshal(reqBody, &printReq); err != nil {
		fmt.Printf("Error: Failed to unmarshal request: %v\n", err)
		jo.SetExit(job.ExitErr)
		return
	}

	fmt.Printf("Parsed message: '%s'\n", printReq.Message)

	jo.Context = &JobContext{
		RequestData: printReq,
	}

	jo.NextStage = "StageProcessData"
	fmt.Println("========================================")
}

// 阶段2: 处理数据
func stageProcessData(jo *job.Job) {
	fmt.Println("\n====== [Stage 2/3: ProcessData] ======")

	ctx, ok := jo.Context.(*JobContext)
	if !ok {
		fmt.Println("Error: Invalid job context type")
		jo.SetExit(job.ExitErr)
		return
	}

	fmt.Printf("Processing message: '%s'\n", ctx.RequestData.Message)
	fmt.Println("Simulating some heavy work for 2 seconds...")
	time.Sleep(2 * time.Second)

	ctx.ProcessTime = time.Now()
	jo.NextStage = "StageFinalize"
	fmt.Println("========================================")
}

// 阶段3: 完成
func stageFinalize(jo *job.Job) {
	fmt.Println("\n====== [Stage 3/3: Finalize] ======")

	ctx, ok := jo.Context.(*JobContext)
	if !ok {
		fmt.Println("Error: Invalid job context type")
		jo.SetExit(job.ExitErr)
		return
	}

	fmt.Printf("Finalizing job for message: '%s'\n", ctx.RequestData.Message)
	fmt.Printf("Data was processed at: %s\n", ctx.ProcessTime.Format(time.RFC3339))
	fmt.Println("Job finished successfully!")

	jo.SetExit(job.ExitSuccess)
	fmt.Println("========================================")
}

// 注册作业状态机
func registerMyJobs() map[string]job.JobStateMachine {
	jsm := job.NewJobStateMachine()
	jsm.InitStage = "StageParseRequest"
	jsm.Async = true
	stageInfo := job.StageInfo{Reentrant: true, Rollbackable: false}

	jsm.AddStage("StageParseRequest", stageParseRequest, stageInfo)
	jsm.AddStage("StageProcessData", stageProcessData, stageInfo)
	jsm.AddStage("StageFinalize", stageFinalize, stageInfo)

	return map[string]job.JobStateMachine{
		"PrintMessage": jsm,
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ops := utils.GetDefaultLogConfig()
	ops.Provider = log.Console
	ops.LogLevel = log.TraceLevel
	logger := utils.GetLogger(ops)

	socketPath := "/tmp/mgr_demo.sock"
	_ = os.Remove(socketPath)
	address := utils.NewNetAddr("unix", socketPath)

	opt := &mgr.Option{
		GroupName: "DemoGroup",
		FuncRegister: map[string]mgr.FuncRegister{
			job.DefaultVersion: registerMyJobs(),
		},
		Address: address,
		Logger:  logger,
	}

	mgrIns := mgr.NewMgr(opt)
	fmt.Println("Server starting, listening on", socketPath)
	mgrIns.Start()

	logger.Flush()
	logger.Stop()
}
