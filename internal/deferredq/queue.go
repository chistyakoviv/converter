package deferredq

import (
	"log/slog"
	"sync"

	"github.com/chistyakoviv/converter/internal/lib/slogger"
)

type callback func() error

type DQueue interface {
	Add(cb ...callback)
	Wait()
	Release()
}

type queue struct {
	done      chan interface{}
	once      sync.Once
	mu        sync.Mutex
	callbacks []callback
	logger    *slog.Logger
}

func New(logger *slog.Logger) DQueue {
	return &queue{
		done:   make(chan interface{}),
		logger: logger,
	}
}

func (q *queue) Add(cb ...callback) {
	q.mu.Lock()
	q.callbacks = append(q.callbacks, cb...)
	q.mu.Unlock()
}

func (q *queue) Wait() {
	<-q.done
}

func (q *queue) Release() {
	q.once.Do(func() {
		defer close(q.done)

		q.mu.Lock()
		// Make a copy of callbacks and clear the original callbacks
		callbacks := q.callbacks
		q.callbacks = nil
		q.mu.Unlock()

		errs := make(chan error, len(callbacks))
		for _, cb := range callbacks {
			// Call all callbacks asyncroniously
			go func(cb callback) {
				errs <- cb()
			}(cb)
		}

		// Block until an error received
		for i := 0; i < cap(errs); i++ {
			if err := <-errs; err != nil {
				q.logger.Error("error from deferred queue", slogger.Err(err))
			}
		}
	})
}
