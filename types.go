package mailman

import (
	"context"
	"time"
)

type Event interface {
	Bind(dst interface{}) error
	Content() []byte
}

type HandlerFunction func(context.Context, Event) error

type MiddlewareFunction func(HandlerFunction) HandlerFunction

type ProcessStatus int

const (
	ProcessStatusIgnore ProcessStatus = iota
	ProcessStatusTimeout
	ProcessStatusSuccess
	ProcessStatusFailure
	ProcessStatusPanic
)

type Router interface {
	Handle() HandlerFunction
	BufferSize() int
	Produce(buffer chan<- Event, signal chan<- struct{})
	Commit(ProcessStatus, Event) error
	Timeout() time.Duration
}
