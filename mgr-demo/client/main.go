package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"code.byted.org/infcs/mgr/kitex_gen/infcs/job/manager"
	// 引入 kitex 生成的客户端代码
	"code.byted.org/infcs/mgr/kitex_gen/infcs/mgr/framework"
	"github.com/cloudwego/kitex/client"
)

// PrintRequest 定义了客户端发送给服务端的数据结构。
// 为了确保序列化和反序列化兼容，此结构体应与服务端定义保持一致。
type PrintRequest struct {
	Message string `json:"message"`
}

func main() {
	// 服务端监听的 Unix Socket 地址，必须与服务端配置完全一致。
	const sockAddr = "/tmp/mgr_demo.sock"
	log.Println("客户端启动，准备向服务端发起 RPC 请求...")

	// 1. 创建一个 Kitex 客户端。
	//    - "mgr-server" 是目标服务的名称，可以自定义，主要用于服务发现和负载均衡。
	//    - client.WithHostPorts 指定了服务端的地址。对于 Unix Socket，地址格式为 "unix:///path/to/socket.sock"。
	c, err := manager.NewClient(
		"mgr-server",
		client.WithHostPorts("unix://"+sockAddr),
	)
	if err != nil {
		log.Fatalf("创建 Kitex 客户端失败: %v", err)
	}

	// 2. 准备要发送的业务数据。
	//    创建一个 PrintRequest 实例，并将其序列化为 JSON 格式的字节流。
	printReq := PrintRequest{
		Message: "你好，MGR！这是一个来自新客户端的任务！",
	}
	reqBytes, err := json.Marshal(printReq)
	if err != nil {
		log.Fatalf("请求数据 JSON 序列化失败: %v", err)
	}

	// 3. 构建 RPC 请求对象 (framework.NewJobReq)。
	//    - JobName 必须与服务端注册的作业名 ("PrintMessage") 完全匹配。
	//    - PrivateReq 字段用于传递自定义的业务数据。
	//    - Timeout 设置了作业的执行超时时间。
	req := &framework.NewJobReq{
		JobName: "PrintMessage", // 作业名，必须与服务端注册的完全一致
		PrivateReq: &framework.PrivateReq{
			ReqBytes: reqBytes, // 包含业务数据的 JSON 字节流
		},
		Timeout: int64(time.Minute.Seconds()), // 设置 1 分钟的超时
	}

	log.Printf("正在发送 NewJob 请求, 作业名: %s", req.JobName)

	// 4. 发起 RPC 调用。
	//    调用客户端的 NewJob 方法，将请求发送到服务端。
	//    context.Background() 用于提供请求的上下文，可用于控制超时和取消。
	resp, err := c.NewJob(context.Background(), req)
	if err != nil {
		log.Fatalf("RPC 调用失败: %v", err)
	}

	// 5. 处理并打印服务端的响应。
	//    如果 RPC 调用成功，服务端会返回一个响应，其中包含作业ID和初始状态。
	log.Printf("成功收到服务端的响应！\n作业ID: %s\n作业状态: %s\n", resp.JobID, resp.Status)
}
