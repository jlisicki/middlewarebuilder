package middlewarebuilder

import "errors"

type (
	Factory[T any] interface {
		Create(next T) (T, error)
	}
	Factories[T any] []Factory[T]

	// Builder builds a middleware chain with a handler as last part of the chain.
	// Since middlewares must be added in a deterministic order, Builder is not thread-safe.
	Builder[T any] struct {
		factories Factories[T]
		handler   *T
	}

	// FactoryFunc implements Factory interface as function.
	FactoryFunc[T any] func(next T) (T, error)
)

func (f FactoryFunc[T]) Create(next T) (T, error) {
	return f(next)
}

func (f Factories[T]) Create(handler T) (T, error) {
	next := handler
	var err error
	for i := len(f) - 1; i >= 0; i-- {
		next, err = f[i].Create(next)
		if err != nil {
			return next, err
		}
	}
	return next, nil
}

var errMissingHandler = errors.New("missing handler")

func NewBuilder[T any]() *Builder[T] {
	return &Builder[T]{}
}

// Add middleware factory. First added middleware is first called in a chain.
func (b *Builder[T]) Add(middlewareFactory Factory[T]) *Builder[T] {
	b.factories = append(b.factories, middlewareFactory)
	return b
}

// WithHandler sets a handler used to build a chain.
func (b *Builder[T]) WithHandler(h T) *Builder[T] {
	b.handler = &h
	return b
}

// Build a chain of middlewares using middleware factories with a handler as last.
func (b *Builder[T]) Build() (T, error) {
	if b.handler == nil {
		var zero T
		return zero, errMissingHandler
	}
	return b.factories.Create(*b.handler)
}
