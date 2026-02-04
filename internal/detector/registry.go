package detector

import (
	"fmt"
	"sync"
)

// Registry manages registered drift detectors for different AWS services.
//
// The registry provides a centralized way to discover and instantiate
// service-specific detectors. Detectors self-register during init().
type Registry struct {
	mu        sync.RWMutex
	factories map[string]DetectorFactory
}

// DefaultRegistry is the global detector registry.
var DefaultRegistry = NewRegistry()

// NewRegistry creates a new detector registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]DetectorFactory),
	}
}

// Register adds a detector factory to the registry.
// Typically called from init() in each detector's package.
func (r *Registry) Register(serviceName string, factory DetectorFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[serviceName] = factory
}

// Get retrieves a detector for the specified service.
// Returns an error if no detector is registered for the service.
func (r *Registry) Get(serviceName string) (Detector, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[serviceName]
	if !ok {
		return nil, fmt.Errorf("no detector registered for service: %s", serviceName)
	}
	return factory(), nil
}

// List returns all registered service names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]string, 0, len(r.factories))
	for name := range r.factories {
		services = append(services, name)
	}
	return services
}

// Has returns true if a detector is registered for the service.
func (r *Registry) Has(serviceName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.factories[serviceName]
	return ok
}

// Register adds a detector factory to the default registry.
func Register(serviceName string, factory DetectorFactory) {
	DefaultRegistry.Register(serviceName, factory)
}

// Get retrieves a detector from the default registry.
func Get(serviceName string) (Detector, error) {
	return DefaultRegistry.Get(serviceName)
}

// List returns all registered service names from the default registry.
func List() []string {
	return DefaultRegistry.List()
}

// Has returns true if a detector is registered in the default registry.
func Has(serviceName string) bool {
	return DefaultRegistry.Has(serviceName)
}
