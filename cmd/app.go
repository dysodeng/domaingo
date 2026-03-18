package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	configCli "github.com/CXeon/tiles/config"
	"github.com/CXeon/tiles/db/gormdb"
	tilesGateway "github.com/CXeon/tiles/gateway"
	"github.com/CXeon/tiles/gateway/traefik"
	tilesLogger "github.com/CXeon/tiles/logger"
	zapLogger "github.com/CXeon/tiles/logger/zap"
	tilesRegistry "github.com/CXeon/tiles/registry"
	"github.com/CXeon/tiles/registry/etcd"
	"github.com/google/uuid"

	apihttp "github.com/CXeon/domaingo/api/http"
	"github.com/CXeon/domaingo/internal/config"
	infraGateway "github.com/CXeon/domaingo/internal/infrastructure/gateway"
	infraRdb "github.com/CXeon/domaingo/internal/infrastructure/rdb"
	infraRegistry "github.com/CXeon/domaingo/internal/infrastructure/registry"
	appLogger "github.com/CXeon/domaingo/internal/logger"
	"github.com/CXeon/domaingo/internal/modular"
	utilip "github.com/CXeon/tiles/util/ip"
)

const stopTimeout = 10 * time.Second

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
	flags  config.Flags
	logger tilesLogger.Logger

	gateway  *infraGateway.Gateway   // nil if disabled
	registry *infraRegistry.Registry // nil if disabled
	rdb      *infraRdb.Rdb
	srv      *apihttp.Server

	stopOnce  sync.Once
	restartCh chan struct{}
}

// New creates a new App from startup flags.
func New(flags config.Flags) *App {
	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		ctx:       ctx,
		cancel:    cancel,
		flags:     flags,
		restartCh: make(chan struct{}, 1),
	}
}

// Init loads configuration and initializes all enabled infrastructure components.
func (a *App) Init() error {
	if err := config.Load(a.flags); err != nil {
		return err
	}

	// 1. Logger (always initialized)
	logCfg := config.Config.Base.Log
	a.logger = zapLogger.NewLogger(zapLogger.Config{
		Filename:     logCfg.Filename,
		MaxSize:      logCfg.MaxSize,
		MaxBackups:   logCfg.MaxBackups,
		MaxAge:       logCfg.MaxAge,
		Compress:     logCfg.Compress,
		Level:        logCfg.Level,
		EnableStdout: logCfg.EnableStdout,
	})
	appLogger.GlobalLogger = a.logger

	// 2. Gateway (if enabled)
	if config.Config.Base.Gateway.Enabled {
		gw, err := a.buildGateway()
		if err != nil {
			return err
		}
		a.gateway = gw
	}

	// 3. Registry (if enabled)
	if config.Config.Base.Registry.Enabled {
		reg, err := a.buildRegistry()
		if err != nil {
			return err
		}
		a.registry = reg
	}

	// 4. Rdb (if main enabled)
	if config.Config.Rdb.Main.Enabled {
		a.rdb = a.buildRdb()
	}

	// 5. HTTP Server
	deps := modular.Deps{
		Rdb:    a.rdb,
		Logger: a.logger,
	}
	srv, err := apihttp.NewServer(
		fmt.Sprintf(":%d", a.flags.HttpPort),
		deps,
		modules(),
	)
	if err != nil {
		return err
	}
	a.srv = srv

	return nil
}

// Start connects all infrastructure clients, starts all services, watches config,
// and blocks until an OS signal or a config-triggered restart is received.
func (a *App) Start() error {
	if a.registry != nil {
		if err := a.registry.Start(a.ctx); err != nil {
			return err
		}
	}

	if a.gateway != nil {
		if err := a.gateway.Start(a.ctx); err != nil {
			return err
		}
	}

	if err := a.rdb.Start(a.ctx); err != nil {
		return err
	}

	go func() {
		if err := a.srv.Start(a.ctx); err != nil {
			a.logger.Error("http server error", err, nil)
		}
	}()

	if err := config.Watch(a.handleConfigChange); err != nil {
		return err
	}

	a.logger.Info("app started", tilesLogger.Fields{"service": a.flags.ServiceName})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		a.logger.Info("received signal, stopping", tilesLogger.Fields{"signal": sig.String()})
		return a.Stop()
	case <-a.restartCh:
		a.logger.Info("config changed, restarting", nil)
		if err := a.Stop(); err != nil {
			a.logger.Error("stop error during restart", err, nil)
		}
		a.restart()
	}
	return nil
}

// Stop gracefully shuts down all resources within stopTimeout.
// Safe to call multiple times; only the first call takes effect.
func (a *App) Stop() (stopErr error) {
	a.stopOnce.Do(func() {
		a.cancel()

		ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
		defer cancel()

		var errs []error

		if a.srv != nil {
			if err := a.srv.Stop(ctx); err != nil {
				errs = append(errs, err)
			}
		}

		if a.gateway != nil {
			if err := a.gateway.Stop(ctx); err != nil {
				errs = append(errs, err)
			}
		}

		if a.registry != nil {
			if err := a.registry.Stop(ctx); err != nil {
				errs = append(errs, err)
			}
		}

		if a.rdb != nil {
			if err := a.rdb.Stop(ctx); err != nil {
				errs = append(errs, err)
			}
		}

		if err := config.Close(); err != nil {
			errs = append(errs, err)
		}

		stopErr = errors.Join(errs...)
	})
	return
}

// handleConfigChange is called by the config watcher on any change.
// Any modification to log, gateway, registry, or rdb config triggers a process restart.
func (a *App) handleConfigChange(event *configCli.ChangeEvent) {
	// Empty Changes slice means a full config reload — always restart.
	needRestart := len(event.Changes) == 0
	if !needRestart {
		for _, change := range event.Changes {
			if strings.HasPrefix(change.Key, "base.log") ||
				strings.HasPrefix(change.Key, "base.gateway") ||
				strings.HasPrefix(change.Key, "base.registry") ||
				strings.HasPrefix(change.Key, "rdb.") {
				needRestart = true
				break
			}
		}
	}
	if needRestart {
		// Non-blocking send; deduplicate rapid consecutive changes.
		select {
		case a.restartCh <- struct{}{}:
		default:
		}
	}
}

// restart replaces the current process with a fresh instance using the same arguments.
// On Linux/macOS, syscall.Exec is used to keep the same PID (required in Docker where
// PID 1 must not exit). On Windows, a new process is spawned and the current one exits.
func (a *App) restart() {
	executable, err := os.Executable()
	if err != nil {
		os.Exit(1)
	}
	if runtime.GOOS == "windows" {
		cmd := exec.Command(executable, os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		_ = cmd.Start()
		os.Exit(0)
	} else {
		// Replaces the current process in-place; the PID is preserved.
		// os.Exit(1) is only reached if Exec itself fails.
		_ = syscall.Exec(executable, os.Args, os.Environ())
		os.Exit(1)
	}
}

func (a *App) buildGateway() (*infraGateway.Gateway, error) {
	gwCfg := config.Config.Base.Gateway

	localIP, err := utilip.GetLocalIP()
	if err != nil {
		return nil, err
	}

	endpoint := &tilesGateway.Endpoint{
		InstanceID: uuid.New().String(),
		Env:        a.flags.Env,
		Cluster:    a.flags.Cluster,
		Company:    a.flags.Company,
		Project:    a.flags.Project,
		Service:    a.flags.ServiceName,
		Protocol:   tilesGateway.ProtocolTypeHttp,
		Color:      a.flags.Color,
		Ip:         localIP,
		Port:       uint16(a.flags.HttpPort),
		TTL:        uint32(gwCfg.TTL),
		Weight:     uint16(gwCfg.Weight),
	}

	provider := &traefik.Provider{
		KVType:         traefik.ProviderTypeEtcd,
		Endpoints:      gwCfg.Provider.Endpoints,
		Username:       gwCfg.Provider.Username,
		Password:       gwCfg.Provider.Password,
		ConnectTimeout: time.Duration(gwCfg.Provider.DialTimeout) * time.Second,
		Namespace:      gwCfg.Provider.Namespace,
	}

	return infraGateway.NewGateway(provider, endpoint), nil
}

func (a *App) buildRegistry() (*infraRegistry.Registry, error) {
	regCfg := config.Config.Base.Registry

	localIP, err := utilip.GetLocalIP()
	if err != nil {
		return nil, err
	}

	endpoint := &tilesRegistry.Endpoint{
		InstanceID: uuid.New().String(),
		Env:        a.flags.Env,
		Cluster:    a.flags.Cluster,
		Company:    a.flags.Company,
		Project:    a.flags.Project,
		Service:    a.flags.ServiceName,
		Protocol:   tilesRegistry.ProtocolTypeHttp,
		Color:      a.flags.Color,
		Ip:         localIP,
		Port:       uint16(a.flags.HttpPort),
		Weight:     uint16(regCfg.Weight),
	}

	cfg := etcd.Config{
		Endpoints:            regCfg.Provider.Endpoints,
		Username:             regCfg.Provider.Username,
		Password:             regCfg.Provider.Password,
		DialTimeout:          time.Duration(regCfg.Provider.DialTimeout) * time.Second,
		LoadBalancerStrategy: regCfg.Provider.LoadBalancerStrategy,
	}

	return infraRegistry.NewRegistry(cfg, endpoint), nil
}

func (a *App) buildRdb() *infraRdb.Rdb {
	rdbCfg := config.Config.Rdb

	mainCfg := gormdb.Config{
		Driver:   rdbCfg.Main.Driver,
		Host:     rdbCfg.Main.Host,
		Port:     rdbCfg.Main.Port,
		Username: rdbCfg.Main.Username,
		Password: rdbCfg.Main.Password,
		Database: rdbCfg.Main.Database,
	}

	var secondaryCfg *gormdb.Config
	if rdbCfg.Secondary.Enabled {
		secondaryCfg = &gormdb.Config{
			Driver:   rdbCfg.Secondary.Driver,
			Host:     rdbCfg.Secondary.Host,
			Port:     rdbCfg.Secondary.Port,
			Username: rdbCfg.Secondary.Username,
			Password: rdbCfg.Secondary.Password,
			Database: rdbCfg.Secondary.Database,
		}
	}

	return infraRdb.NewRdb(mainCfg, secondaryCfg)
}
