package client

import "context"

type Client interface {
	GetName() string
	IsEnabled(ctx context.Context) bool
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
}
