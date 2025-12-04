package generic

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/guilhermealvess/mailman"
)

type GenericEvent[T any] struct {
	msg   *T
	raw   []byte
	retry chan struct{}
	count int
}

func NewGenericEvent[T any](obj T) (*GenericEvent[T], error) {
	raw, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return &GenericEvent[T]{
		msg:   &obj,
		raw:   raw,
		retry: make(chan struct{}),
		count: 1,
	}, nil
}

func (ge *GenericEvent[T]) Bind(dst interface{}) error {
	if err := json.Unmarshal(ge.raw, dst); err != nil {
		return fmt.Errorf("generic router: %v", err)
	}

	return nil
}

func (ge *GenericEvent[T]) Content() []byte {
	return ge.raw
}

type GenericRouter[T any] struct {
	handler    mailman.HandlerFunction
	size       int
	timeout    time.Duration
	publisher  chan T
	maxRetries int
}

func NewGenericRouter[T any](fn mailman.HandlerFunction) (mailman.Router, chan T) {
	pub := make(chan T)
	return &GenericRouter[T]{
		handler:    fn,
		size:       10,
		timeout:    time.Second * 30,
		publisher:  pub,
		maxRetries: 5,
	}, pub
}

func (gr *GenericRouter[T]) Handle() mailman.HandlerFunction {
	return gr.handler
}

func (gr *GenericRouter[T]) BufferSize() int {
	return gr.size
}

func (gr *GenericRouter[T]) Produce(buffer chan<- mailman.Event, signal chan<- struct{}) {
	for msg := range gr.publisher {
		event, _ := NewGenericEvent(msg)
		buffer <- event

		go func() {
			for range event.retry {
				if event.count > gr.maxRetries {
					return
				}
				buffer <- event
			}
		}()
	}

	signal <- struct{}{}
}

func (gr *GenericRouter[T]) Commit(status mailman.ProcessStatus, event mailman.Event) error {
	iEvent, ok := event.(*GenericEvent[T])
	if !ok {
		return errors.New("TODO")
	}

	iEvent.count++

	switch status {
	case mailman.ProcessStatusFailure, mailman.ProcessStatusPanic, mailman.ProcessStatusTimeout:
		iEvent.retry <- struct{}{}

	case mailman.ProcessStatusIgnore, mailman.ProcessStatusSuccess:
		close(iEvent.retry)
	}

	return nil
}

func (gr *GenericRouter[T]) Timeout() time.Duration {
	return gr.timeout
}
