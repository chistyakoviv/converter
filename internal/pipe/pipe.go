package pipe

import "net/http"

type HandlerFn[T any] func(deps *T, fn http.HandlerFunc) http.HandlerFunc

type Pipe[T any] interface {
	Pipe(n HandlerFn[T]) Pipe[T]
	Build() http.HandlerFunc
}

type pipe[T any] struct {
	deps  *T
	funcs []HandlerFn[T]
}

func New[T any](deps *T) Pipe[T] {
	return &pipe[T]{
		deps: deps,
	}
}

func (p *pipe[T]) Pipe(fn HandlerFn[T]) Pipe[T] {
	p.funcs = append(p.funcs, fn)
	return p
}

func (p *pipe[T]) Build() http.HandlerFunc {
	var next http.HandlerFunc
	for i := len(p.funcs) - 1; i > 0; i-- {
		next = p.funcs[i](p.deps, next)
	}
	return next
}
