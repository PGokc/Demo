//go:build example

package client

// 说明：
// 该示例展示基于 Kitex 的 Action 调用方式。要编译运行本文件，需：
// 1) 在字节内网准备好对应的 Kitex IDL，并生成 client 代码与 kitex.yml
// 2) 使用 -tags example 进行编译或运行（避免默认构建集成该示例）
// 3) 服务端默认监听 :8888，可通过 callopt.WithHostPort("127.0.0.1:8888") 指定

import (
	"context"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/google/uuid"
)

// MockAction 展示一个 Action 的异步调用示例
func MockAction() error {
	ctx := context.Background()

	// 以下 Request 类型与构造函数由 Kitex 根据 IDL 生成，请替换为你本地生成的代码
	// var req = &yourkitex.Request{}
	// req.MgrReq.Action = "Deploy"
	// req.MgrReq.JobID = uuid.NewString()
	// req.MgrReq.Async = true
	// req.MgrReq.Stage = "Stage1"

	// 示例客户端创建，具体以你生成的 Kitex 客户端 API 为准
	cli, err := client.NewClient("anything")
	if err != nil {
		return err
	}

	// resp, err := cli.Action(ctx, req, callopt.WithHostPort("127.0.0.1:8888"))
	// if err != nil {
	// 	return err
	// }
	_ = callopt.WithHostPort
	_ = uuid.NewString

	return nil
}
