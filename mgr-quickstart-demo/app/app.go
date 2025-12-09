package app

import (
	"fmt"

	"code.byted.org/infcs/mgr/pkg/job"
)

// App 演示私有实例的定义与在 Stage 中的使用
// 生产环境中可注入日志、RPC 封装、DAO 等
// 通过 job.App 获取并在业务阶段内调用

type App struct {
	Name string
}

func (a App) Install(j *job.Job) {
	fmt.Printf("[App] %s installing in stage %s\n", a.Name, j.GetCurStage())
}
