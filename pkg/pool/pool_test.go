package pool

import (
	"sync"
	"testing"
)

// mockResetter is a mock implementation of Resetter that tracks Reset() calls
type mockResetter struct {
	resetCalled bool
	resetCount  int
	mu          sync.Mutex
}

// Reset marks that Reset was called
func (m *mockResetter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resetCalled = true
	m.resetCount++
}

// wasResetCalled returns true if Reset was called
func (m *mockResetter) wasResetCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.resetCalled
}

// getResetCount returns the number of times Reset was called
func (m *mockResetter) getResetCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.resetCount
}

// newMockResetter creates a new mockResetter (needed for Pool factory)
func newMockResetter() *mockResetter {
	return &mockResetter{}
}

func TestPool_GetPut(t *testing.T) {
	p := New(func() *mockResetter {
		return &mockResetter{}
	})

	// Get an object
	obj := p.Get()

	// Verify Reset hasn't been called yet
	if obj.wasResetCalled() {
		t.Error("expected Reset to not be called before Put")
	}

	// Put it back - this should call Reset
	p.Put(obj)

	// Verify Reset was called
	if !obj.wasResetCalled() {
		t.Error("expected Reset to be called after Put")
	}

	// Get again - should be same object (or equivalent) and reset should not be called on Get
	obj2 := p.Get()

	// Verify the object was reset before being put back
	if !obj2.wasResetCalled() {
		t.Error("expected object to have been reset after Put")
	}

	// Verify Reset was only called once
	if obj2.getResetCount() != 1 {
		t.Errorf("expected Reset to be called 1 time, got %d", obj2.getResetCount())
	}
}

func TestPool_ResetCalledOnPut(t *testing.T) {
	// Test with mock to verify Reset is called
	p := New(newMockResetter)

	obj := p.Get()

	// Put should call Reset
	p.Put(obj)

	// Verify it was reset
	if !obj.wasResetCalled() {
		t.Error("expected Reset to be called after Put")
	}

	if obj.getResetCount() != 1 {
		t.Errorf("expected Reset to be called 1 time, got %d", obj.getResetCount())
	}
}

func TestPool_ResetCalledMultipleTimes(t *testing.T) {
	// Test that Reset is called each time Put is called
	p := New(newMockResetter)

	obj := p.Get()

	// Put it back - first reset
	p.Put(obj)

	// Get it again
	obj2 := p.Get()

	// Modify and put back - second reset
	p.Put(obj2)

	// Get it once more
	obj3 := p.Get()

	// Verify Reset was called twice (once for each Put)
	if obj3.getResetCount() != 2 {
		t.Errorf("expected Reset to be called 2 times, got %d", obj3.getResetCount())
	}
}

func TestPool_ConcurrentAccess(t *testing.T) {
	p := New(newMockResetter)

	const numGoroutines = 100
	const iterationsPerGoroutine = 100

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterationsPerGoroutine; j++ {
				obj := p.Get()

				// Return to pool
				p.Put(obj)

				// Verify Reset was called
				if !obj.wasResetCalled() {
					t.Errorf("Reset not called in goroutine %d, iteration %d", id, j)
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestPool_MultiplePools(t *testing.T) {
	p1 := New(newMockResetter)
	p2 := New(newMockResetter)

	obj1 := p1.Get()
	p1.Put(obj1)

	obj2 := p2.Get()
	p2.Put(obj2)

	// Get from each pool and verify Reset was called
	result1 := p1.Get()
	if result1 == nil {
		t.Error("expected non-nil object from p1")
	}
	if !result1.wasResetCalled() {
		t.Error("expected Reset to be called for p1")
	}

	result2 := p2.Get()
	if result2 == nil {
		t.Error("expected non-nil object from p2")
	}
	if !result2.wasResetCalled() {
		t.Error("expected Reset to be called for p2")
	}

	// Verify objects are independent (different Reset counts)
	result1.resetCount = 5
	result2.resetCount = 10

	if result1.getResetCount() != 5 {
		t.Errorf("p1 object should have 5 resets, got %d", result1.getResetCount())
	}
	if result2.getResetCount() != 10 {
		t.Errorf("p2 object should have 10 resets, got %d", result2.getResetCount())
	}
}

func TestPool_CustomFactory(t *testing.T) {
	// Test that custom factory function is used
	callCount := 0

	p := New(func() *mockResetter {
		callCount++
		return &mockResetter{}
	})

	// Get an object - should call factory
	obj1 := p.Get()

	if callCount != 1 {
		t.Errorf("expected factory to be called 1 time, got %d", callCount)
	}

	// Put it back
	p.Put(obj1)

	// Get again - should reuse object, factory not called
	obj2 := p.Get()

	if callCount != 1 {
		t.Errorf("expected factory to still be called 1 time (object reused), got %d", callCount)
	}

	// Verify Reset was called
	if !obj2.wasResetCalled() {
		t.Error("expected Reset to be called after Put")
	}

	// Get a third time (while second is still in use)
	_ = p.Get()

	if callCount != 2 {
		t.Errorf("expected factory to be called 2 times, got %d", callCount)
	}
}

func TestPool_NilReceiver(t *testing.T) {
	// Test that Pool handles nil receivers safely
	var p *Pool[*mockResetter]

	// These should panic or handle gracefully
	defer func() {
		if r := recover(); r != nil {
			// Panic is expected for nil pool
		}
	}()

	_ = p.Get()
}

func TestPool_GetPutWithNilValue(t *testing.T) {
	// Test that the pool doesn't accept nil values
	p := New(func() *mockResetter {
		return &mockResetter{}
	})

	obj := p.Get()

	// Try to put nil - this should work but won't call Reset
	p.Put(nil)

	// Verify the object we got is still valid
	if obj == nil {
		t.Error("expected non-nil object from Get")
	}
}

func TestPool_MultiplePutGet(t *testing.T) {
	p := New(newMockResetter)

	// Put multiple objects
	var objects []*mockResetter
	for range 10 {
		obj := p.Get()
		p.Put(obj)
		objects = append(objects, obj)
	}

	// Get them back - should reuse objects
	var resetCount int
	for range 10 {
		obj := p.Get()
		resetCount += obj.getResetCount()
		p.Put(obj)
	}

	// All objects should have been reset at least once
	if resetCount < 10 {
		t.Errorf("expected at least 10 resets, got %d", resetCount)
	}
}

func TestPool_PutSameObjectMultipleTimes(t *testing.T) {
	p := New(newMockResetter)

	obj := p.Get()

	// Put same object multiple times
	p.Put(obj)
	count1 := obj.getResetCount()

	p.Put(obj)
	count2 := obj.getResetCount()

	// Reset should be called each time
	if count2 != count1+1 {
		t.Errorf("expected Reset count to increase by 1, got %d vs %d", count2, count1)
	}
}

func BenchmarkPool_GetPut(b *testing.B) {
	p := New(newMockResetter)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := p.Get()
			p.Put(obj)
		}
	})
}

func BenchmarkPool_NoPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := &mockResetter{}
			obj.Reset()
			// No pool, object is just garbage collected
		}
	})
}
