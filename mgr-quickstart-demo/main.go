package main

import (
	"fmt"
	"os"
	"time"

	apppkg "mgr-quickstart-demo/app"
	demo "mgr-quickstart-demo/demo"

	"code.byted.org/infcs/mgr/pkg/job"
	"code.byted.org/infcs/mgr/pkg/mgr"
	"code.byted.org/infcs/mgr/pkg/utils"
)

func main() {
	// 构造业务 App 实例并注入到 mgr.Option
	app := apppkg.App{Name: "PGtest-Quickstart-App"}

	opt := &mgr.Option{
		// 可由命令行覆盖
		Id:             "",
		ElectionFlag:   false,
		ZkAddress:      []string{},
		SessionTimeout: 6 * time.Second, // 不指定时默认 6s

		App:       app, // 注入业务 App 实例
		GroupName: "PGtestMgrQuickstart",
		FuncRegister: map[string]mgr.FuncRegister{
			job.DefaultVersion: demo.RegisterJobFunc(),
		},
		RegisterFuncInNewMgr: true,
		// 监听 :8888
		Address: utils.NewNetAddr("tcp", ":8888"),
		// 指定数据库连接信息
		//JobServiceOption: &mgr.JobServiceOption{
		//	JobMode: service.JobModeLocal,
		//	JobDBOption: &dao.DBOption{
		//		UserName: "root",
		//		Password: "123456",
		//		Host:     "127.0.0.1",
		//		Port:     "3306",
		//		DBName:   "mgr",
		//	},
		//},
	}

	InitMgrOps(opt, os.Args)
	fmt.Printf("Starting mgr on %s, Id=%s, Group=%s, Election=%v\n", ":8888", opt.Id, opt.GroupName, opt.ElectionFlag)
	mgr.NewAndStart(opt)
}

var help = func() {
	fmt.Println("Usage for pgtest mgr quickstart")
	fmt.Println("====================================================")
	fmt.Println("Single replica (no election):")
	fmt.Println("  go run main.go --id s1 --electionFlag 0 --groupName MgrQuickstart --groupId 123")
	fmt.Println()
	fmt.Println("Three replicas with election (ZK required):")
	fmt.Println("  go run main.go --id s1 --electionFlag 1 --zkAdress 10.227.31.8:2181 --groupName MgrQuickstart --groupId 123")
	fmt.Println("  go run main.go --id s2 --electionFlag 1 --zkAdress 10.227.31.8:2181 --groupName MgrQuickstart --groupId 123")
	fmt.Println("  go run main.go --id s3 --electionFlag 1 --zkAdress 10.227.31.8:2181 --groupName MgrQuickstart --groupId 123")
	fmt.Println()
	fmt.Println("Optional:")
	fmt.Println("  --sessionTimeout 8  # seconds")
}

// InitMgrOps 解析命令行参数并填充 mgr.Option
func InitMgrOps(ops *mgr.Option, args []string) {
	if len(args) < 2 || args == nil {
		help()
		return
	}

	for i := 1; i < len(args); i++ {
		key := args[i]
		if i+1 >= len(args) {
			break
		}
		val := args[i+1]

		switch key {
		case "--id":
			ops.Id = val
			i++
		case "--electionFlag":
			if val == "1" || val == "true" {
				ops.ElectionFlag = true
			} else {
				ops.ElectionFlag = false
			}
			i++
		case "--zkAdress", "--zkAddress":
			ops.ZkAddress = append(ops.ZkAddress, val)
			i++
		case "--groupName":
			ops.GroupName = val
			i++
		case "--sessionTimeout":
			// 以秒为单位解析
			if d, err := time.ParseDuration(val + "s"); err == nil {
				ops.SessionTimeout = d
			}
			i++
		default:
			// ignore unknown
		}
	}
}
