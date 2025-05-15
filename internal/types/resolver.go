package types

import (
	"fmt"
	"sync"

	"github.com/google/wire"
)

// GenericResolver provides a type-safe way to resolve implementations of interfaces based on keys
type GenericResolver[K comparable, V any] struct {
	implementations map[K]V
	mu              sync.RWMutex
}

// NewGenericResolver creates a new generic resolver
func NewGenericResolver[K comparable, V any]() *GenericResolver[K, V] {
	return &GenericResolver[K, V]{
		implementations: make(map[K]V),
	}
}

// Register adds an implementation to the resolver
func (r *GenericResolver[K, V]) Register(key K, implementation V) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.implementations[key] = implementation
}

// Resolve returns an implementation for a given key
func (r *GenericResolver[K, V]) Resolve(key K) (V, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if implementation, ok := r.implementations[key]; ok {
		return implementation, nil
	}

	var zero V
	return zero, fmt.Errorf("no implementation found for key: %v", key)
}

// MustResolve returns an implementation for a given key or panics if not found
func (r *GenericResolver[K, V]) MustResolve(key K) V {
	impl, err := r.Resolve(key)
	if err != nil {
		panic(err)
	}
	return impl
}

// Implementations returns all registered implementations
func (r *GenericResolver[K, V]) Implementations() map[K]V {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to avoid mutation issues
	result := make(map[K]V, len(r.implementations))
	for k, v := range r.implementations {
		result[k] = v
	}

	return result
}

// Implementation is a helper type for registering implementations with Wire
type Implementation[K comparable, V any] struct {
	Key            K
	Implementation V
}

// NewImplementation creates a new implementation entry
func NewImplementation[K comparable, V any](key K, implementation V) Implementation[K, V] {
	return Implementation[K, V]{
		Key:            key,
		Implementation: implementation,
	}
}

// RegisterImplementation registers an implementation with a resolver
func RegisterImplementation[K comparable, V any](resolver *GenericResolver[K, V], impl Implementation[K, V]) {
	resolver.Register(impl.Key, impl.Implementation)
}

// ProvideResolver creates a new resolver with the provided implementations
func ProvideResolver[K comparable, V any](impls []Implementation[K, V]) *GenericResolver[K, V] {
	resolver := NewGenericResolver[K, V]()
	for _, impl := range impls {
		resolver.Register(impl.Key, impl.Implementation)
	}
	return resolver
}

// ResolverSet provides a way to create a resolver in Wire
var ResolverSet = wire.NewSet(
// This is empty as specific resolver types will need their own wire sets
)
