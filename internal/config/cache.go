package config

import (
	"context"

	configCli "github.com/CXeon/tiles/config"
	"github.com/CXeon/tiles/config/apollo"
	"github.com/CXeon/tiles/config/viper"
)

const cacheLocalFileName = "cache"

type cacheConn struct {
	Addr         string `json:"addr" mapstructure:"addr"`
	Password     string `json:"password" mapstructure:"password"`
	DB           int    `json:"db" mapstructure:"db"`
	PoolSize     int    `json:"pool_size" mapstructure:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns" mapstructure:"min_idle_conns"`
	DialTimeout  int    `json:"dial_timeout" mapstructure:"dial_timeout"`   // seconds
	ReadTimeout  int    `json:"read_timeout" mapstructure:"read_timeout"`   // seconds
	WriteTimeout int    `json:"write_timeout" mapstructure:"write_timeout"` // seconds
}

type cacheConfig struct {
	TokenStore cacheConn `json:"token_store" mapstructure:"token_store"`
}

type CacheLoader struct {
	cli configCli.Config
}

func NewCacheLoader(opt Option) *CacheLoader {
	var cli configCli.Config
	if opt.f.ConfigMode == Remote {
		cfg := apollo.Config{
			AppID:          opt.f.ServiceName,
			Cluster:        opt.f.Cluster,
			IP:             opt.remote.IP,
			NamespaceName:  opt.remote.NamespaceName,
			Secret:         opt.remote.Secret,
			IsBackupConfig: true,
		}
		cli = apollo.New(cfg)
	}
	if opt.f.ConfigMode != Remote {
		cfg := viper.Config{
			ConfigPaths: opt.local.paths,
			ConfigName:  cacheLocalFileName,
			ConfigType:  opt.local.typ,
			EnvPrefix:   opt.local.envPrefix,
			AutoEnv:     opt.local.autoEnv,
		}
		cli = viper.New(cfg)
	}
	return &CacheLoader{cli: cli}
}

func (l *CacheLoader) Load() (*cacheConfig, error) {
	if err := l.cli.Load(); err != nil {
		return nil, err
	}
	return l.Unmarshal()
}

func (l *CacheLoader) Unmarshal() (*cacheConfig, error) {
	var cfg struct {
		Cache cacheConfig `mapstructure:"cache"`
	}
	if err := l.cli.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg.Cache, nil
}

func (l *CacheLoader) Watch(handler configCli.ChangeHandler) error {
	return l.cli.Watch(handler)
}

func (l *CacheLoader) Close() error {
	return l.cli.Close(context.Background())
}
