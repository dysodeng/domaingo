package cmd

import (
	"context"

	"github.com/CXeon/domaingo/internal/config"
	"github.com/CXeon/domaingo/internal/infrastructure/client"
	"github.com/CXeon/domaingo/internal/infrastructure/runner"
)

// Flags 包装程序启动参数，在程序的生命周期内不允许修改
// type Flags struct {
// 	ConfigMode  string // 配置模式 local 加载本地配置文件，remote 从远程配置中心加载
// 	Env         string // 环境名称
// 	Cluster     string // 集群名称
// 	ServiceName string // 服务名称
// 	Color       string // 染色
// 	HttpPort    uint   // 服务监听端口
// }

type App struct {
	ctx    context.Context
	flags  config.Flags
	client map[string]client.Client // 基础设施客户端
	runner map[string]runner.Runner // 基础设施持续运行的服务
}
