package pool

import (
	"fmt"
	"sync"
	"testing"
)

// TestStruct is a simple test struct with exported fields
type TestStruct struct {
	ID     int
	Name   string
	Items  []string
	Config map[string]string
}

// Reset resets TestStruct to its zero values
func (t *TestStruct) Reset() {
	t.ID = 0
	t.Name = ""
	t.Items = t.Items[:0]
	clear(t.Config)
}

// TestStructWithNested is a test struct with nested fields
type TestStructWithNested struct {
	Value  int
	Nested *TestStruct
	Slice  []*TestStruct
}

// Reset resets TestStructWithNested to its zero values
func (t *TestStructWithNested) Reset() {
	t.Value = 0
	if t.Nested != nil {
		t.Nested.Reset()
	}
	t.Slice = t.Slice[:0]
}

func TestPool_GetPut(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			Config: make(map[string]string),
		}
	})

	// Get an object
	obj := p.Get()

	// Modify it
	obj.ID = 42
	obj.Name = "test"
	obj.Items = []string{"item1", "item2", "item3"}
	obj.Config = map[string]string{"key": "value"}

	// Put it back
	p.Put(obj)

	// Get again - should be same object (or equivalent) and reset
	obj2 := p.Get()

	// Verify it was reset
	if obj2.ID != 0 {
		t.Errorf("expected ID to be 0 after Put/Get, got %d", obj2.ID)
	}

	if obj2.Name != "" {
		t.Errorf("expected Name to be empty after Put/Get, got %s", obj2.Name)
	}

	if len(obj2.Items) != 0 {
		t.Errorf("expected Items to have length 0 after Put/Get, got %d", len(obj2.Items))
	}

	if len(obj2.Config) != 0 {
		t.Errorf("expected Config to be empty after Put/Get, got %d items", len(obj2.Config))
	}
}

func TestPool_ResetCalledOnPut(t *testing.T) {
	// Test with TestStruct which already has Reset method
	p := New(func() *TestStruct {
		return &TestStruct{
			Config: make(map[string]string),
		}
	})

	obj := p.Get()
	obj.ID = 100

	// Put should call Reset
	p.Put(obj)

	// Verify it was reset (we can check ID is 0)
	if obj.ID != 0 {
		t.Errorf("expected ID to be 0 after Put, got %d", obj.ID)
	}
}

// Test with TestStructWithNested to verify nested reset
func TestPool_ResetNestedStructs(t *testing.T) {
	p := New(func() *TestStructWithNested {
		return &TestStructWithNested{}
	})

	obj := p.Get()
	obj.Value = 100
	obj.Nested = &TestStruct{ID: 1, Name: "nested"}
	obj.Slice = []*TestStruct{{ID: 2}, {ID: 3}}

	p.Put(obj)

	// Verify nested struct was reset
	if obj.Nested == nil {
		t.Error("expected Nested to be non-nil after Reset")
	} else {
		if obj.Nested.ID != 0 {
			t.Errorf("expected Nested.ID to be 0 after Reset, got %d", obj.Nested.ID)
		}
		if obj.Nested.Name != "" {
			t.Errorf("expected Nested.Name to be empty after Reset, got %s", obj.Nested.Name)
		}
	}

	if len(obj.Slice) != 0 {
		t.Errorf("expected Slice to have length 0 after Reset, got %d", len(obj.Slice))
	}
}

func TestPool_ConcurrentAccess(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			Items:  make([]string, 0, 10),
			Config: make(map[string]string),
		}
	})

	const numGoroutines = 100
	const iterationsPerGoroutine = 100

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterationsPerGoroutine; j++ {
				obj := p.Get()

				// Modify object
				obj.ID = id*100 + j
				obj.Name = "test"
				obj.Items = append(obj.Items, fmt.Sprintf("item-%d", j))
				obj.Config[fmt.Sprintf("key-%d", j)] = "value"

				// Return to pool
				p.Put(obj)

				// Verify it's reset
				if obj.ID != 0 || obj.Name != "" || len(obj.Items) != 0 || len(obj.Config) != 0 {
					t.Errorf("object not properly reset after Put")
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestPool_MultiplePools(t *testing.T) {
	p1 := New(func() *TestStruct {
		return &TestStruct{}
	})

	p2 := New(func() *TestStructWithNested {
		return &TestStructWithNested{}
	})

	obj1 := p1.Get()
	obj1.ID = 42
	p1.Put(obj1)

	obj2 := p2.Get()
	obj2.Value = 100
	p2.Put(obj2)

	// Get from each pool and verify types
	result1 := p1.Get()
	if result1 == nil {
		t.Error("expected non-nil TestStruct from p1")
	}

	result2 := p2.Get()
	if result2 == nil {
		t.Error("expected non-nil TestStructWithNested from p2")
	}
}

func TestPool_CustomFactory(t *testing.T) {
	initialCapacity := 1000

	p := New(func() *TestStruct {
		return &TestStruct{
			Items:  make([]string, 0, initialCapacity),
			Config: make(map[string]string),
		}
	})

	obj1 := p.Get()
	capacity1 := cap(obj1.Items)

	if capacity1 != initialCapacity {
		t.Errorf("expected initial capacity %d, got %d", initialCapacity, capacity1)
	}

	// Fill slice
	for i := 0; i < 100; i++ {
		obj1.Items = append(obj1.Items, fmt.Sprintf("item-%d", i))
	}

	// Put back and get again
	p.Put(obj1)

	obj2 := p.Get()
	capacity2 := cap(obj2.Items)

	if capacity2 != initialCapacity {
		t.Errorf("expected preserved capacity %d, got %d", initialCapacity, capacity2)
	}

	// Verify length is 0 after Reset
	if len(obj2.Items) != 0 {
		t.Errorf("expected length 0 after Reset, got %d", len(obj2.Items))
	}
}

// TestPool_NilReceiver removed - cannot safely call Reset() on nil pointers

func TestPool_MemoryReuse(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			Items:  make([]string, 0, 100),
			Config: make(map[string]string),
		}
	})

	// First usage
	obj1 := p.Get()
	obj1Addr := obj1
	obj1.Items = append(obj1.Items, "item1", "item2", "item3")
	obj1.Config["key"] = "value"

	p.Put(obj1)

	// Second usage - should reuse same object
	obj2 := p.Get()

	// Check if it's same object (not guaranteed, but likely)
	if obj2 == obj1Addr {
		// Verify capacity is preserved
		if cap(obj2.Items) != 100 {
			t.Errorf("expected capacity 100, got %d", cap(obj2.Items))
		}
	}

	// Verify it's reset
	if len(obj2.Items) != 0 {
		t.Errorf("expected empty slice after Reset, got %d elements", len(obj2.Items))
	}

	if len(obj2.Config) != 0 {
		t.Errorf("expected empty map after Reset, got %d elements", len(obj2.Config))
	}
}

func TestPool_NestedStructs(t *testing.T) {
	p := New(func() *TestStructWithNested {
		return &TestStructWithNested{}
	})

	obj := p.Get()

	// Set nested struct
	obj.Nested = &TestStruct{
		ID:   42,
		Name: "test",
	}
	obj.Slice = []*TestStruct{{ID: 1}, {ID: 2}}

	p.Put(obj)

	// Get again
	obj2 := p.Get()

	// Verify nested struct was reset
	if obj2.Nested != nil {
		if obj2.Nested.ID != 0 {
			t.Errorf("expected Nested.ID to be 0, got %d", obj2.Nested.ID)
		}
		if obj2.Nested.Name != "" {
			t.Errorf("expected Nested.Name to be empty, got %s", obj2.Nested.Name)
		}
	}

	if len(obj2.Slice) != 0 {
		t.Errorf("expected Slice to be empty after Reset, got %d elements", len(obj2.Slice))
	}
}

func TestPool_SliceOfPointers(t *testing.T) {
	p := New(func() *TestStructWithNested {
		return &TestStructWithNested{}
	})

	obj := p.Get()

	// Add pointers to slice
	obj.Slice = append(obj.Slice,
		&TestStruct{ID: 1, Name: "item1"},
		&TestStruct{ID: 2, Name: "item2"},
	)

	p.Put(obj)

	obj2 := p.Get()

	// Verify slice is empty but capacity preserved
	if len(obj2.Slice) != 0 {
		t.Errorf("expected slice to be empty, got %d elements", len(obj2.Slice))
	}
}

func TestPool_MultiplePutGet(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{}
	})

	// Put multiple objects
	for i := 0; i < 10; i++ {
		obj := p.Get()
		obj.ID = i
		p.Put(obj)
	}

	// Get them back
	for i := 0; i < 10; i++ {
		obj := p.Get()
		if obj.ID != 0 {
			t.Errorf("expected ID to be 0, got %d", obj.ID)
		}
		p.Put(obj)
	}
}

func TestPool_PointerFields(t *testing.T) {
	p := New(func() *TestStructWithNested {
		return &TestStructWithNested{}
	})

	obj := p.Get()
	obj.Nested = &TestStruct{ID: 100, Name: "test"}
	obj.Slice = []*TestStruct{{ID: 200}}

	p.Put(obj)

	obj2 := p.Get()

	// Verify pointer fields are still not nil (they were allocated)
	if obj2.Nested == nil {
		t.Error("expected Nested to be non-nil")
	} else {
		if obj2.Nested.ID != 0 {
			t.Errorf("expected Nested.ID to be 0, got %d", obj2.Nested.ID)
		}
		if obj2.Nested.Name != "" {
			t.Errorf("expected Nested.Name to be empty, got %s", obj2.Nested.Name)
		}
	}
}

func TestPool_MapFields(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{
			Config: make(map[string]string),
		}
	})

	obj := p.Get()
	obj.Config["key1"] = "value1"
	obj.Config["key2"] = "value2"

	p.Put(obj)

	obj2 := p.Get()

	// Verify map is cleared
	if len(obj2.Config) != 0 {
		t.Errorf("expected Config to be empty, got %d elements", len(obj2.Config))
	}

	// Map should still be non-nil (clear() keeps it)
	if obj2.Config == nil {
		t.Error("expected Config to be non-nil after clear")
	}
}

func BenchmarkPool_GetPut(b *testing.B) {
	p := New(func() *TestStruct {
		return &TestStruct{
			Items:  make([]string, 0, 10),
			Config: make(map[string]string),
		}
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := p.Get()
			obj.ID = 42
			obj.Items = append(obj.Items, "item1", "item2", "item3")
			p.Put(obj)
		}
	})
}

func BenchmarkPool_NoPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := &TestStruct{
				Items:  make([]string, 0, 10),
				Config: make(map[string]string),
			}
			obj.ID = 42
			obj.Items = append(obj.Items, "item1", "item2", "item3")
			// No pool, object is just garbage collected
		}
	})
}
