package main

import (
	"fmt"
	"time"

	"mgr-quickstart-demo/app"
	"mgr-quickstart-demo/demo"

	log "code.byted.org/infcs/lib-log"
	"code.byted.org/infcs/mgr/kitex_gen/infcs/mgr/framework/appservice"
	"code.byted.org/infcs/mgr/pkg/mgr"
	"code.byted.org/infcs/mgr/pkg/utils"
	kitexserver "code.byted.org/kite/kitex/server"
)

func main() {
	// 0. 构造业务 App 实例并注入到 mgr.Option
	app := app.App{Name: "PGtest-Quickstart-App"}

	// 1. 初始化日志
	ops := log.Ops{
		Path:     "./mgr-demo.log",
		Provider: log.StorageProvider(2),
		Prefixs:  []string{"[mgr-demo]"},
		LogLevel: int(log.TraceLevel),
	}
	logger := log.Init(ops)
	log.StartLogger()
	log.SetLevel(int(log.InfoLevel)) //默认Log级别Info
	defer func() {
		log.Flush()
		log.Stop()
	}()

	opt := &mgr.Option{
		// 可由命令行覆盖
		Id:             "",
		ElectionFlag:   false,
		ZkAddress:      []string{},
		SessionTimeout: 6 * time.Second, // 不指定时默认 6s

		App:          app, // 注入业务 App 实例
		GroupName:    "PGtestMgrQuickstart",
		FuncRegister: demo.RegisterJobFunc(),
		// 监听 :8889
		Address: utils.NewNetAddr("tcp", ":8889"),
		Logger:  logger,
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

	MgrIns := mgr.NewMgr(opt)
	MgrIns.Start()

	// Start RPC server
	AppImpl := &mgr.AppServiceImpl{
		MgrIns: MgrIns,
	}

	if opt.Address == nil {
		opt.Address = utils.DefaultAddress
	}
	kLogger := &mgr.KLogger{Logger: logger}
	svr := appservice.NewServer(AppImpl, kitexserver.WithServiceAddr(opt.Address))
	err := svr.Run()
	if err != nil {
		fmt.Printf("Failed to start server: %v", err)
		return
	}
	kLogger.Logger.Info("Starting mgr on %s, Id=%s, Group=%s, Election=%v\n", opt.Address, opt.Id, opt.GroupName, opt.ElectionFlag)
}
