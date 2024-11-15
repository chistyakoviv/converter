package di

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// Thread safe dependency injection container
type Container interface {
	Register(name string, constructor interface{})
	RegisterSingleton(name string, constructor interface{})
	resolve(name string) (interface{}, error)
	resolveWithTracking(name string, resolving map[string]bool) (interface{}, error)
	Has(name string) bool
}

// Container struct
type container struct {
	services   map[string]reflect.Value
	singletons map[string]interface{}
	mu         sync.Mutex
}

// NewContainer creates a new Container instance
func NewContainer() *container {
	return &container{
		services:   make(map[string]reflect.Value),
		singletons: make(map[string]interface{}),
	}
}

// Register registers a service with a constructor function
func (c *container) Register(name string, constructor interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[name] = reflect.ValueOf(constructor)
}

// RegisterSingleton registers a singleton service
func (c *container) RegisterSingleton(name string, constructor interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services[name] = reflect.ValueOf(constructor)
	c.singletons[name] = nil // Placeholder to indicate this is a singleton
}

// Has checks if a service is registered
func (c *container) Has(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.services[name]
	return ok
}

// Resolve resolves a registered service and returns an interface
func (c *container) resolve(name string) (interface{}, error) {
	return c.resolveWithTracking(name, make(map[string]bool))
}

// resolveWithTracking resolves a service with dependency tracking to prevent circular dependencies
func (c *container) resolveWithTracking(name string, resolving map[string]bool) (interface{}, error) {
	c.mu.Lock()

	// Check if the service exists
	constructor, ok := c.services[name]
	if !ok {
		c.mu.Unlock()
		return nil, errors.New("service " + name + " not registered")
	}

	// Check if it's a singleton and already created
	if instance, ok := c.singletons[name]; ok && instance != nil {
		c.mu.Unlock()
		return instance, nil
	}

	// Check for circular dependency
	if resolving[name] {
		c.mu.Unlock()
		return nil, errors.New("circular dependency detected while resolving " + name)
	}

	// Mark this dependency as being resolved
	resolving[name] = true
	c.mu.Unlock() // Unlock after marking to allow other threads to resolve different services

	// Resolve the constructor, passing the container as an argument
	// If the constructor calls other services, they will be resolved without issues
	// because a new tracking map is created for each resolution.
	result := constructor.Call([]reflect.Value{reflect.ValueOf(c)})

	// If it's a singleton, store the created instance
	c.mu.Lock()
	if _, isSingleton := c.singletons[name]; isSingleton {
		// Check if another thread has already resolved this singleton
		if c.singletons[name] == nil {
			c.singletons[name] = result[0].Interface()
		} else {
			// Another thread resolved it; discard the redundant result
			result[0] = reflect.ValueOf(c.singletons[name])
		}
	}
	c.mu.Unlock()

	// Mark this dependency as resolved
	delete(resolving, name)

	return result[0].Interface(), nil
}

// Resolve is a helper function that attempts to resolve and cast a service to the expected type
func Resolve[T any](c Container, name string) (T, error) {
	resolved, err := c.resolve(name)
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
