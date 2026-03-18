package rdb

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/CXeon/tiles/db/gormdb"
	gormlib "gorm.io/gorm"
)

// Rdb wraps a primary and optional secondary gormdb client, providing read/write separation.
type Rdb struct {
	mainCfg      gormdb.Config
	secondaryCfg *gormdb.Config
	opts         []gormdb.Option

	mainCli      *gormdb.Client
	secondaryCli *gormdb.Client

	running atomic.Bool
}

func NewRdb(mainCfg gormdb.Config, secondaryCfg *gormdb.Config, opts ...gormdb.Option) *Rdb {
	return &Rdb{
		mainCfg:      mainCfg,
		secondaryCfg: secondaryCfg,
		opts:         opts,
	}
}

// Start opens connections to main and secondary databases and verifies connectivity via ping.
func (r *Rdb) Start(ctx context.Context) error {
	mainCli, err := gormdb.New(r.mainCfg, r.opts...)
	if err != nil {
		return err
	}
	if err = ping(ctx, mainCli); err != nil {
		_ = mainCli.Close()
		return err
	}
	r.mainCli = mainCli

	if r.secondaryCfg != nil {
		secondaryCli, err := gormdb.New(*r.secondaryCfg, r.opts...)
		if err != nil {
			_ = r.mainCli.Close()
			return err
		}
		if err = ping(ctx, secondaryCli); err != nil {
			_ = r.mainCli.Close()
			_ = secondaryCli.Close()
			return err
		}
		r.secondaryCli = secondaryCli
	}

	r.running.Store(true)
	return nil
}

// Stop closes all database connections.
func (r *Rdb) Stop(_ context.Context) error {
	if !r.running.Swap(false) {
		return nil
	}
	var errs []error
	if r.mainCli != nil {
		if err := r.mainCli.Close(); err != nil {
			errs = append(errs, err)
		}
		r.mainCli = nil
	}
	if r.secondaryCli != nil {
		if err := r.secondaryCli.Close(); err != nil {
			errs = append(errs, err)
		}
		r.secondaryCli = nil
	}
	return errors.Join(errs...)
}

// Writer returns the primary database instance for write operations.
func (r *Rdb) Writer() *gormlib.DB {
	return r.mainCli.GetDB()
}

// Reader returns the secondary database instance for read operations.
// Falls back to the primary if no secondary is configured.
func (r *Rdb) Reader() *gormlib.DB {
	if r.secondaryCli != nil {
		return r.secondaryCli.GetDB()
	}
	return r.mainCli.GetDB()
}

func ping(ctx context.Context, cli *gormdb.Client) error {
	pool, err := cli.Pool()
	if err != nil {
		return err
	}
	return pool.PingContext(ctx)
}
