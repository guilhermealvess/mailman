package mailman

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
	"time"
)

var mu sync.Mutex

var logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{}))

type manager struct {
	routers     sync.Map
	middlewares []MiddlewareFunction
}

func New() *manager {
	return &manager{}
}

func (m *manager) Register(name string, router Router) {
	m.routers.Store(name, router)
}

func (m *manager) Use(middleware MiddlewareFunction) {
	m.middlewares = append(m.middlewares, middleware)
}

func (m *manager) Run() {
	if !mu.TryLock() {
		log.Fatal("mailman: manager is already running")
	}
	defer mu.Unlock()

	signal := make(chan struct{})

	m.routers.Range(func(key, value any) bool {
		var (
			name       = key.(string)
			router     = value.(Router)
			timeout    = router.Timeout()
			bufferSize = router.BufferSize()
			buffer     = make(chan Event, bufferSize)
		)

		fn := router.Handle()
		for _, middleware := range m.middlewares {
			fn = middleware(fn)
		}

		for idx := range bufferSize {
			pid := fmt.Sprintf("%s-%d", name, idx)
			go router.Produce(buffer, signal)

			go func() {
				for event := range buffer {
					ctx := context.Background()
					ctx, cancel := context.WithTimeout(ctx, timeout)
					ctx = WithContext(ctx, Context{
						HandlerName: name,
						PID:         pid,
						StartAt:     time.Now(),
					})

					done := make(chan ProcessStatus, 1)
					go func() {
						defer func() {
							if r := recover(); r != nil {
								done <- ProcessStatusPanic
							}
						}()

						logger.InfoContext(ctx, "Start process")

						if err := fn(ctx, event); err != nil {
							done <- ProcessStatusFailure
							return
						}

						done <- ProcessStatusSuccess
					}()

					var status ProcessStatus
					select {
					case <-ctx.Done():
						status = ProcessStatusTimeout

					case status = <-done:
						// println("TODO", status)
					}

					if err := router.Commit(status, event); err != nil {
						signal <- struct{}{}
					}

					close(done)
					cancel()
				}
			}()
		}

		return true
	})

	<-signal
}
