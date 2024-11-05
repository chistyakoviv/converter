package di

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

type Container interface {
	Context() context.Context
	Register(name string, constructor interface{})
	RegisterSingleton(name string, constructor interface{})
	Resolve(name string) (interface{}, error)
	Has(name string) bool
}

// Container struct
type container struct {
	ctx        context.Context
	services   map[string]reflect.Value
	singletons map[string]interface{}
	// For now the implementation is not thread safe,
	// because nested dependency resolving causes a deadlock
	// mutex      sync.Mutex
}

// NewContainer creates a new Container instance
func NewContainer(ctx context.Context) *container {
	return &container{
		ctx:        ctx,
		services:   make(map[string]reflect.Value),
		singletons: make(map[string]interface{}),
	}
}

// Context returns the context
func (c *container) Context() context.Context {
	return c.ctx
}

// Register registers a service with a constructor function
func (c *container) Register(name string, constructor interface{}) {
	// c.mutex.Lock()
	// defer c.mutex.Unlock()

	c.services[name] = reflect.ValueOf(constructor)
}

// RegisterSingleton registers a singleton service
func (c *container) RegisterSingleton(name string, constructor interface{}) {
	// c.mutex.Lock()
	// defer c.mutex.Unlock()

	c.services[name] = reflect.ValueOf(constructor)
	c.singletons[name] = nil // Placeholder to indicate this is a singleton
}

// Has checks if a service is registered
func (c *container) Has(name string) bool {
	// c.mutex.Lock()
	// defer c.mutex.Unlock()

	_, ok := c.services[name]
	return ok
}

// Resolve resolves a registered service and returns an interface
func (c *container) Resolve(name string) (interface{}, error) {
	// c.mutex.Lock()
	// defer c.mutex.Unlock()

	// Check if the service exists
	constructor, ok := c.services[name]
	if !ok {
		return nil, errors.New("service " + name + " not registered")
	}

	// Handle singletons
	if instance, ok := c.singletons[name]; ok && instance != nil {
		return instance, nil
	} else if ok && instance == nil {
		// Create singleton instance with container passed as argument
		result := constructor.Call([]reflect.Value{reflect.ValueOf(c)})
		c.singletons[name] = result[0].Interface()
		return c.singletons[name], nil
	}

	// Call constructor for non-singleton service with container passed as argument
	result := constructor.Call([]reflect.Value{reflect.ValueOf(c)})
	return result[0].Interface(), nil
}

// Resolve is a helper function that attempts to resolve and cast a service to the expected type
func Resolve[T any](c Container, name string) (T, error) {
	resolved, err := c.Resolve(name)
	if err != nil {
		var zero T
		return zero, err
	}
	casted, ok := resolved.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("failed to cast resolved service to expected type")
	}
	return casted, nil
}
