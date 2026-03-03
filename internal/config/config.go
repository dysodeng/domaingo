package config

import (
	"errors"
	"sync"

	configCli "github.com/CXeon/tiles/config"
)

var Config *config
var once sync.Once

type config struct {
	l          sync.RWMutex
	baseLoader *BaseLoader

	Base *base `json:"base"`
}

type Option struct {
	mode   Mode
	local  *LocalConfig
	remote *RemoteConfig
}

// 设置配置模式
func WithMode(mode Mode) func(opt *Option) {
	return func(opt *Option) {
		opt.mode = mode
	}
}

// 设置远程配置参数
func WithRemote(remote *RemoteConfig) func(opt *Option) {
	return func(opt *Option) {
		opt.remote = remote
	}
}

type Mode string

const (
	Local  Mode = "local"
	Remote Mode = "remote"
)

type LocalConfig struct {
	Paths []string `json:"paths"`
	// Name      string   `json:"name"`
	Type      string `json:"type"`
	EnvPrefix string `json:"env_prefix"`
	AutoEnv   bool   `json:"auto_env"`
}

type RemoteConfig struct {
	AppID         string `json:"app_id"`
	Cluster       string `json:"cluster"`
	IP            string `json:"ip"`
	NamespaceName string `json:"namespace_name"`
	Secret        string `json:"secret"`
	IsBackup      bool   `json:"is_backup"`
}

func Load(opts ...func(opt *Option)) error {
	// 默认加载本地配置
	defaultOpt := Option{
		mode: Local,
		local: &LocalConfig{
			Paths:     []string{"./configs/"},
			Type:      "yaml",
			EnvPrefix: "DOMAINGO",
			AutoEnv:   true,
		},
		remote: nil,
	}

	for _, opt := range opts {
		opt(&defaultOpt)
	}

	if defaultOpt.mode == Remote && defaultOpt.remote == nil {
		return errors.New("remote config must required in remote mode")
	}

	once.Do(func() {
		Config = &config{
			l: sync.RWMutex{},
		}

		// 加载base配置
		baseLoader := NewBaseLoader(defaultOpt)
		base, err := baseLoader.Load()
		if err != nil {
			panic(err)
		}
		Config.baseLoader = baseLoader
		Config.Base = base

	})

	return nil
}

func Watch(handler configCli.ChangeHandler) error {

	if Config == nil || Config.baseLoader == nil {
		return errors.New("config not loaded, call Load() first")
	}

	var err error
	err = Config.baseLoader.Watch(func(event *configCli.ChangeEvent) {
		// 1. 自动同步全局 Config.Base
		Config.l.Lock()
		newBase, err := Config.baseLoader.Unmarshal()
		if err == nil {
			Config.Base = newBase
		}
		Config.l.Unlock()

		// 2. 将变更事件传给调用方，由调用方决定副作用
		if handler != nil {
			handler(event)
		}
	})

	return err
}
