package main

import (
	"context"
	"fmt"
	"log"
	"mgr-quickstart-demo/demo"
	"time"

	"code.byted.org/infcs/mgr/kitex_gen/infcs/mgr/framework"
	"code.byted.org/infcs/mgr/kitex_gen/infcs/mgr/framework/appservice"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/google/uuid"
)

/*
*
Deploy Action：这是一个异步 Job (Async = true)。这意味着当我们调用它时，服务端会立即返回一个响应，而真正的部署工作（Stage1 和 Stage2）会在后台执行。
执行流程：Deploy Job 首先进入 Stage1，模拟工作 1 秒后，自动转换到 Stage2，再工作 1 秒后，整个 Job 标记为成功。
GetDeployStatus Action：这是一个同步 Job，它的唯一作用就是根据请求中提供的 GetJobStatusID 来查询另一个 Job 的当前状态（CurStatus）和所处阶段（CurStage），并将其返回。
这正是 mgr 框架典型的“发起异步任务 -> 轮询任务状态”的工作模式。
*/
func main() {
	// 1. 创建客户端
	cli, err := appservice.NewClient("PGtest-Mgr-Demo", client.WithHostPorts("127.0.0.1:8889"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 2. 发起 "Deploy" 异步 Job
	jobID, err := startDeployJob(cli)
	if err != nil {
		log.Fatalf("Failed to start deploy job: %v", err)
	}
	if jobID == "" {
		log.Fatalf("Started job but did not get a JobID")
	}
	log.Printf("✅ Successfully started an async 'Deploy' job, JobID: %s\n", jobID)

	// 3. 轮询 Job 状态直到其完成或失败
	log.Println("----------------------------------------------------")
	log.Println("🔍 Start polling for job status...")
	pollJobStatus(cli, jobID)
}

// startDeployJob 发起一个 Deploy Job 并返回其 JobID
func startDeployJob(cli appservice.Client) (string, error) {
	req := &framework.Request{
		MgrReq: &framework.MgrReq{
			Ctx: &framework.JobContext{
				Action:    demo.DeployAction,
				RequestID: uuid.NewString(),
				Product:   framework.Product_RDS,
			},
		},
	}

	log.Println("🚀 Sending 'Deploy' request to start an async job...")
	resp, err := cli.Action(context.Background(), req, callopt.WithRPCTimeout(3*time.Second))
	if err != nil {
		return "", fmt.Errorf("RPC call failed: %w", err)
	}
	log.Printf("✅ Received mgr framework response: %+v", resp)

	// 从初始响应中提取 GetJobStatusID，这是后续查询状态的凭证
	log.Printf("JobId:%s", resp.MgrResp.Ctx.RequestID)
	return resp.MgrResp.Ctx.RequestID, nil
}

// pollJobStatus 循环查询指定 JobID 的状态，并打印 Stage 变化
func pollJobStatus(cli appservice.Client, jobID string) {
	// 轮询最多 10 秒
	timeout := time.After(15 * time.Second)
	// 每 500 毫秒查询一次
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			log.Println("❌ Polling timed out after 10 seconds.")
			return
		case <-ticker.C:
			req := &framework.Request{
				MgrReq: &framework.MgrReq{
					Ctx: &framework.JobContext{
						Product:        framework.Product_RDS,
						JobOps:         framework.JobOption_Init,
						Action:         demo.GetDeployStatusAction, // 使用专用于查询状态的 Action
						RequestID:      uuid.NewString(),
						GetJobStatusID: jobID, // 传入要查询的 JobID
					},
				},
			}

			resp, err := cli.Action(context.Background(), req, callopt.WithRPCTimeout(3*time.Second))
			if err != nil {
				log.Printf("⚠️ Error polling status: %v", err)
				continue
			}
			log.Printf("✅ Received mgr framework response: %+v", resp)

			status := resp.MgrResp.Ctx.GetCurStatus()
			stage := resp.MgrResp.Ctx.GetCurStage()
			log.Printf("  -> Polling... JobID: [%s], Current Status: [%s], Current Stage: [%s]", jobID, status, stage)

			// Job 完成或失败，则退出轮询
			if status == framework.JobStatus_Completed || status == framework.JobStatus_Failed {
				log.Printf("✅ Job [%s] finished with final status: [%s]", jobID, status)
				log.Println("----------------------------------------------------")
				return
			}
		}
	}
}
