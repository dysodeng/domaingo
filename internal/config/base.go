package config

import (
	configCli "github.com/CXeon/tiles/config"
	"github.com/CXeon/tiles/config/apollo"
	"github.com/CXeon/tiles/config/viper"
)

const baseLocalFileName = "base"

type base struct {
	Name    string `json:"name" mapstructure:"name"`
	Env     string `json:"env" mapstructure:"env"`
	Cluster string `json:"cluster" mapstructure:"cluster"`
	Company string `json:"company" mapstructure:"company"`
	Project string `json:"project" mapstructure:"project"`
	Color   string `json:"color" mapstructure:"color"`
	Server  struct {
		Http string `json:"http" mapstructure:"http"`
		Port int    `json:"port" mapstructure:"port"`
	} `json:"server" mapstructure:"server"`
	Log struct {
		Filename     string `json:"filename" mapstructure:"filename"`
		Level        string `json:"level" mapstructure:"level"`
		MaxSize      int    `json:"max_size" mapstructure:"max_size"`
		MaxBackups   int    `json:"max_backups" mapstructure:"max_backups"`
		MaxAge       int    `json:"max_age" mapstructure:"max_age"`
		Compress     bool   `json:"compress" mapstructure:"compress"`
		EnableStdout bool   `json:"enable_stdout" mapstructure:"enable_stdout"`
	} `json:"log" mapstructure:"log"`
	Gateway struct {
		Enabled bool   `json:"enabled" mapstructure:"enabled"`
		Ip      string `json:"ip" mapstructure:"ip"`
		Port    int    `json:"port" mapstructure:"port"`
		TTL     int    `json:"ttl" mapstructure:"ttl"`
		Weight  int    `json:"weight" mapstructure:"weight"`
	} `json:"gateway" mapstructure:"gateway"`
	Registry struct {
		Enabled bool   `json:"enabled" mapstructure:"enabled"`
		Ip      string `json:"ip" mapstructure:"ip"`
		Port    int    `json:"port" mapstructure:"port"`
		Weight  int    `json:"weight" mapstructure:"weight"`
	} `json:"registry" mapstructure:"registry"`
}

type BaseLoader struct {
	cli configCli.Config
}

func NewBaseLoader(opt Option) *BaseLoader {
	var cli configCli.Config
	if opt.mode == Remote {
		cfg := apollo.Config{
			AppID:          opt.remote.AppID,
			Cluster:        opt.remote.Cluster,
			IP:             opt.remote.IP,
			NamespaceName:  opt.remote.NamespaceName,
			Secret:         opt.remote.Secret,
			IsBackupConfig: true,
		}
		cli = apollo.New(cfg)
	}
	if opt.mode != Remote {
		cfg := viper.Config{
			ConfigPaths: opt.local.Paths,
			ConfigName:  baseLocalFileName,
			ConfigType:  opt.local.Type,
			EnvPrefix:   opt.local.EnvPrefix,
			AutoEnv:     opt.local.AutoEnv,
		}
		cli = viper.New(cfg)
	}
	return &BaseLoader{cli: cli}
}

func (l *BaseLoader) Load() (*base, error) {
	var cfg base
	err := l.cli.Load()
	if err != nil {
		return nil, err
	}
	err = l.cli.UnmarshalKey(baseLocalFileName, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (l *BaseLoader) Unmarshal() (*base, error) {
	var cfg base
	if err := l.cli.UnmarshalKey(baseLocalFileName, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (l *BaseLoader) Watch(handler configCli.ChangeHandler) error {
	return l.cli.Watch(handler)
}
