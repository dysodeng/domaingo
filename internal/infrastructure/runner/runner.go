package runner

import "context"

type Runner interface {
	GetName() string
	IsEnabled(ctx context.Context) bool
	IsRunning(ctx context.Context) bool
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
