package cloudflarewebhook

import (
	"context"
	"time"
)

type PostInput struct {
	Binding string
	Path    string
	Timeout time.Duration
	Headers map[string]string
	Body    []byte
}

type Bridge interface {
	PostJSON(ctx context.Context, input PostInput) error
}
