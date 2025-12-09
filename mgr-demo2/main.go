package main

import (
	"flag"
	"fmt"
	"strings"

	log "code.byted.org/infcs/lib-log"
	"code.byted.org/infcs/mgr/pkg/mgr"
	"code.byted.org/infcs/mgr/pkg/utils"

	"mgr-demo2/app"
	"mgr-demo2/demo"
)

// 支持 --zkAddress 多次或逗号分隔的解析
// 例如：--zkAddress=zk1:2181,zk2:2181 或 --zkAddress=zk1:2181 --zkAddress=zk2:2181
type sliceFlag []string

func (s *sliceFlag) String() string { return strings.Join(*s, ",") }
func (s *sliceFlag) Set(v string) error {
	if v == "" {
		return nil
	}
	for _, p := range strings.Split(v, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			*s = append(*s, p)
		}
	}
	return nil
}

func main() {
	ops := log.Ops{
		Path:     "./mgr-demo2.log",
		Provider: log.StorageProvider(2),
		Prefixs:  []string{"[mgr-demo2]"},
		LogLevel: log.TraceLevel,
	}
	logger := log.Init(ops)
	log.StartLogger()
	log.SetLevel(log.InfoLevel) //默认Log级别Info
	defer func() {
		log.Flush()
		log.Stop()
	}()

	var (
		id           = flag.String("id", "demo-1", "节点唯一标识，用于 leader 选举")
		electionFlag = flag.Bool("electionFlag", false, "是否启用 ZK 选主")
		groupName    = flag.String("groupName", "MgrQuickStart", "产品分组名，用于构造选举目录和区分产品")
		zkAddresses  sliceFlag
	)
	flag.Var(&zkAddresses, "zkAddress", "ZK 地址，支持多次或逗号分隔，如: host1:2181,host2:2181")
	flag.Parse()

	// 构造 AppIns 注入示例
	appIns := &app.App{ // App 实现了 job.AppIns 的空接口
		Name: "quickstart-app",
		Meta: map[string]string{
			"env":  "dev",
			"demo": "mgr-quickstart",
		},
	}

	// 注册 Action 的状态机
	funcRegister := demo.Register() // map[string]job.JobStateMachine

	// 组装 mgr.Option
	opt := &mgr.Option{
		Id:           *id,
		ElectionFlag: *electionFlag,
		ZkAddress:    zkAddresses,
		GroupName:    *groupName,
		// Address 可使用 tcp 或默认
		Address:      utils.DefaultAddress, // 等价于 utils.NewNetAddr("tcp", ":8889")
		App:          appIns,
		FuncRegister: funcRegister,
		Logger:       logger,
	}

	// 打印启动参数，便于学习观察
	fmt.Printf("Start Mgr with params: id=%s electionFlag=%v groupName=%s zk=%v addr=%v\n",
		opt.Id, opt.ElectionFlag, opt.GroupName, opt.ZkAddress, opt.Address)

	// 启动服务（包含选主、JobManager、Kitex Server）
	mgr.NewAndStart(opt)
}
