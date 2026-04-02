package config

import (
	"context"

	configCli "github.com/CXeon/tiles/config"
	"github.com/CXeon/tiles/config/apollo"
	"github.com/CXeon/tiles/config/viper"
)

const rdbLocalFileName = "rdb"

type rdbConn struct {
	Enabled     bool   `json:"enabled" mapstructure:"enabled"`
	AutoMigrate bool   `json:"auto_migrate" mapstructure:"auto_migrate"`
	Driver      string `json:"driver" mapstructure:"driver"`
	Host        string `json:"host" mapstructure:"host"`
	Port        int    `json:"port" mapstructure:"port"`
	Username    string `json:"username" mapstructure:"username"`
	Password    string `json:"password" mapstructure:"password"`
	Database    string `json:"database" mapstructure:"database"`
}

type rdb struct {
	Main      rdbConn `json:"main" mapstructure:"main"`
	Secondary rdbConn `json:"secondary" mapstructure:"secondary"`
}

type RdbLoader struct {
	cli configCli.Config
}

func NewRdbLoader(opt Option) *RdbLoader {
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
			ConfigName:  rdbLocalFileName,
			ConfigType:  opt.local.typ,
			EnvPrefix:   opt.local.envPrefix,
			AutoEnv:     opt.local.autoEnv,
		}
		cli = viper.New(cfg)
	}
	return &RdbLoader{cli: cli}
}

func (l *RdbLoader) Load() (*rdb, error) {

	err := l.cli.Load()
	if err != nil {
		return nil, err
	}
	rdb, err := l.Unmarshal()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}

func (l *RdbLoader) Unmarshal() (*rdb, error) {
	var cfg struct {
		Rdb rdb `mapstructure:"rdb"`
	}
	if err := l.cli.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg.Rdb, nil
}

func (l *RdbLoader) Watch(handler configCli.ChangeHandler) error {
	return l.cli.Watch(handler)
}

func (l *RdbLoader) Close() error {
	return l.cli.Close(context.Background())
}
