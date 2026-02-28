package cmd

import (
	"context"

	"github.com/CXeon/domaingo/internal/infrastructure/client"
	"github.com/CXeon/domaingo/internal/infrastructure/runner"
	"github.com/CXeon/tiles/config"
)

type Flags struct {
	ConfigMode string // 配置模式 local 加载本地配置文件，remote 从远程配置中心加载
	Env        string // 环境名称
	Cluster    string // 集群名称
	// Color      string // 染色
	Port int // 服务监听端口
}

type App struct {
	ctx    context.Context
	flags  Flags
	cfg    config.Config
	client []client.Client // 基础设施客户端
	runner []runner.Runner // 基础设施持续运行的服务
}
