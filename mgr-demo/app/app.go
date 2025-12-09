package app

import "code.byted.org/infcs/mgr/pkg/job"

// App 是一个轻量的业务实例结构体，实现 job.AppIns 空接口，用于注入到 Mgr。
// 在 Stage 中可以通过 jo.App 获取并打印信息。
type App struct {
	Name string
	Meta map[string]string
}

// 断言满足 job.AppIns
var _ job.AppIns = (*App)(nil)
