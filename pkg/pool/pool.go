package pool

import (
	"sync"
)

// Resetter is an interface that types must implement to be used with Pool.
type Resetter interface {
	Reset()
}

// Pool is a generic pool that stores objects of type T.
// Type T must implement the Resetter interface.
//
// Pool automatically resets objects before returning them to the pool.
type Pool[T Resetter] struct {
	pool sync.Pool
}

// New creates a new Pool[T] with the provided function to create new objects.
// The newFunc is called when the pool is empty and Get is called.
//
// Example:
//
//	type MyStruct struct { ... }
//	func (m *MyStruct) Reset() { ... }
//
//	p := pool.New(func() *MyStruct {
//	    return &MyStruct{}
//	})
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
	}
}

// Get retrieves an object from the pool.
// If the pool is empty, a new object is created using the newFunc provided to New.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns an object to the pool.
// The object's Reset method is called before it is placed in the pool.
//
// Example:
//
//	obj := p.Get()
//	// use obj...
//	p.Put(obj)
func (p *Pool[T]) Put(x T) {
	x.Reset()
	p.pool.Put(x)
}
