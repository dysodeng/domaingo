package registry

import (
	"context"
	"errors"
	"sync/atomic"

	tilesRegistry "github.com/CXeon/tiles/registry"
	"github.com/CXeon/tiles/registry/etcd"
)

// Registry wraps an etcd registry client, managing endpoint registration and lifecycle.
type Registry struct {
	cfg      etcd.Config
	endpoint *tilesRegistry.Endpoint
	registry *etcd.Registry
	running  atomic.Bool
}

func NewRegistry(cfg etcd.Config, endpoint *tilesRegistry.Endpoint) *Registry {
	return &Registry{
		cfg:      cfg,
		endpoint: endpoint,
	}
}

// Start establishes a connection to etcd and registers the service endpoint.
func (r *Registry) Start(ctx context.Context) error {
	reg, err := etcd.NewRegistry(r.cfg)
	if err != nil {
		return err
	}
	if err = reg.Register(ctx, r.endpoint); err != nil {
		_ = reg.Close(ctx)
		return err
	}
	r.registry = reg
	r.running.Store(true)
	return nil
}

// Stop deregisters the service endpoint and closes the etcd connection.
func (r *Registry) Stop(ctx context.Context) error {
	if !r.running.Swap(false) {
		return nil
	}
	if r.registry == nil {
		return nil
	}
	var errs []error
	if err := r.registry.Deregister(ctx, r.endpoint); err != nil {
		errs = append(errs, err)
	}
	if err := r.registry.Close(ctx); err != nil {
		errs = append(errs, err)
	}
	r.registry = nil
	return errors.Join(errs...)
}
