package mailman

import (
	"context"
	"time"
)

type Context struct {
	HandlerName string
	PID         string
	StartAt     time.Time
}

type contentKey string

const _contextKey = contentKey("MailmanContext")

func WithContext(ctx context.Context, value Context) context.Context {
	return context.WithValue(ctx, _contextKey, value)
}
