// client/main.go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/google/uuid"

	"mgr-demo2/kitex_gen/thrift/kitex/api"
	"mgr-demo2/kitex_gen/thrift/kitex/api/appservice"
)

func main() {
	// 1. 创建客户端
	cli, err := appservice.NewClient("AppService", client.WithHostPorts("127.0.0.1:8889"))
	if err != nil {
		panic(err)
	}

	// 2. 构造请求
	req := &api.MgrReq{
		Action: "Deploy",
		JobID:  uuid.NewString(),
		Async:  true, // 异步执行
	}

	// 3. 发起调用
	resp, err := cli.Action(context.Background(), req, callopt.WithRPCTimeout(3*time.Second))
	if err != nil {
		fmt.Printf("RPC call failed: %v\n", err)
		return
	}

	fmt.Printf("Response: %+v\n", resp)
}
