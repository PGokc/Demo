package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"k8s.io/client-go/tools/clientcmd"

	"code.byted.org/larkarch/operation-framework/apis/v1alpha1"
	"code.byted.org/larkarch/operation-framework/pkg/clientset/ops"
	"code.byted.org/larkarch/operation-framework/pkg/framework"
)

func main() {
	// 0. 初始化K8s客户端
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/bytedance/.kube/config")
	if err != nil {
		panic(fmt.Errorf("failed to load kubeconfig: %v", err))
	}
	cli := ops.NewForConfigOrDie(config)

	// 1. 初始化Operation Framework框架
	f := framework.NewFramework(cli, v1alpha1.CompDemo)

	// 2. 注册加法操作
	framework.RegisterExecutor[*AddInput, *AddOutput](f, &Executor{}, &AddInput{})

	// 3. 注册乘法操作
	framework.RegisterExecutor[*MultiplyInput, *MultiplyOutput](f, &MultiplyExecutor{}, &MultiplyInput{})

	// 4. 注册RDS分库操作
	framework.RegisterExecutor[*RdsSplitTest, *framework.EmptyParams](f, &RdsExecutor{}, &RdsSplitTest{})

	// 5. 启动Operation Framework框架
	go f.Start(context.Background())

	// 6. 阻塞以持续运行
	select {}
}

// 1. 定义加法操作
type AddInput struct {
	A float32
	B float32
}

type AddOutput struct {
	C float32
}
type Executor struct{}

func (e *Executor) PreCheck(ctx context.Context, in *AddInput) error {
	return nil
}

func (e *Executor) Name() string { return "add" }
func (e *Executor) Execute(ctx context.Context, in *AddInput) (output *AddOutput, err error) {
	log.Printf("add %v + %v", in.A, in.B)
	return &AddOutput{C: float32(in.A + in.B)}, nil
}
func (e *Executor) PostCheck(ctx context.Context, in *AddInput) error { return nil }

// 2. 定义乘法操作
type MultiplyInput struct {
	A float32
	B float32
}

type MultiplyOutput struct {
	C float32
}

type MultiplyExecutor struct{}

func (e *MultiplyExecutor) PreCheck(ctx context.Context, in *MultiplyInput) error {
	return nil
}

func (e *MultiplyExecutor) Name() string { return "multiply" }
func (e *MultiplyExecutor) Execute(ctx context.Context, in *MultiplyInput) (output *MultiplyOutput, err error) {
	log.Printf("multiply %v * %v", in.A, in.B)
	return &MultiplyOutput{C: float32(in.A * in.B)}, nil
}
func (e *MultiplyExecutor) PostCheck(ctx context.Context, in *MultiplyInput) error { return nil }

// 3. 定义RDS分库操作
type RdsSplitTest struct {
	InstanceId string `json:"instanceId,omitempty"`
	DbName     string `json:"dbName,omitempty"`
	Number     int32  `json:"number,omitempty"`
}

type RdsExecutor struct{}

func (e *RdsExecutor) PreCheck(ctx context.Context, in *RdsSplitTest) error {
	return nil
}
func (e *RdsExecutor) Name() string { return "rds_split" }
func (e *RdsExecutor) Execute(ctx context.Context, in *RdsSplitTest) (output *framework.EmptyParams, err error) {
	log.Printf("rds split %v %v %v", in.InstanceId, in.DbName, in.Number)
	if strings.Contains(in.InstanceId, "bdgate") {
		return nil, fmt.Errorf("instance id contains bdgate")
	}
	return &framework.EmptyParams{}, nil
}
