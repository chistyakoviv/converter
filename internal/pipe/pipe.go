package pipe

type HandlerFn[T any, P any] func(deps *T, fn P) P

type Pipe[T any, P any] interface {
	Pipe(fn HandlerFn[T, P]) Pipe[T, P]
	Build() P
}

type pipe[T any, P any] struct {
	deps  *T
	funcs []HandlerFn[T, P]
}

func New[T any, P any](deps *T) Pipe[T, P] {
	return &pipe[T, P]{
		deps: deps,
	}
}

func (p *pipe[T, P]) Pipe(fn HandlerFn[T, P]) Pipe[T, P] {
	p.funcs = append(p.funcs, fn)
	return p
}

func (p *pipe[T, P]) Build() P {
	var next P
	for i := len(p.funcs) - 1; i >= 0; i-- {
		// Pass deps by ref to avoid unnecessary copyings of heavy objects
		next = p.funcs[i](p.deps, next)
	}
	return next
}
