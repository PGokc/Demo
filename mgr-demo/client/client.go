package main

import (
	"context"
	"log"
	"mgr-demo2/demo"
	"time"

	// 直接从 mgr 库中导入预生成的 api 包 (用于数据结构)
	"code.byted.org/infcs/mgr/kitex_gen/infcs/mgr/framework"
	// 直接从 mgr 库中导入预生成的 appservice 包 (用于客户端)
	"code.byted.org/infcs/mgr/kitex_gen/infcs/mgr/framework/appservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/client/callopt"
	"github.com/google/uuid"
)

/**
1. mgr 是一个库：通过 go.mod 引入 code.byted.org/infcs/mgr，它是一个包含了预编译代码和预生成 Kitex 代码的 Go 模块（库）。
2. 无需本地生成：既然这个库已经提供了与 mgr 框架内部服务配套的、现成的 kitex_gen 代码，我们完全不需要在自己的项目里再放一个 IDL 文件，也不需要再运行 kitex 命令来生成代码。那样做反而会因为版本不匹配而导致冲突。
3. 直接使用库中的代码：正确的做法就是直接 import 并使用 mgr 库中提供的客户端代码和数据结构。
这完美地解释了之前所有的问题。我们一直在试图用我们本地的“蓝图”（IDL）去匹配一个已经建好的“大楼”（mgr 框架），而您现在发现，这座“大楼”的开发商已经把官方的、精确的“访客指南”（预生成的客户端代码）直接提供给我们了。
*/

func mockRpcCall() {
	// 1. 使用从 mgr 库导入的 NewClient 创建客户端
	cli, err := appservice.NewClient("PGtest-Mgr-Demo2", client.WithHostPorts("127.0.0.1:8889"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 2. 使用从 mgr 库导入的数据结构来构造请求
	// 这与我们之前推断的结构完全一致，现在我们是直接使用官方定义
	req := &framework.Request{
		MgrReq: &framework.MgrReq{
			Ctx: &framework.JobContext{
				Action:    demo.DeployAction, // 这个 Action 字符串需要与您在 demo/register.go 中注册的键匹配
				RequestID: uuid.NewString(),
				Product:   framework.Product_RDS,
			},
		},
	}

	log.Printf("Sending request using official mgr library client: %+v\n", req)

	// 3. 发起 RPC 调用
	resp, err := cli.Action(context.Background(), req, callopt.WithRPCTimeout(3*time.Second))
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}

	// 4. 打印成功的响应
	log.Printf("Successfully received response: %+v\n", resp)
}
