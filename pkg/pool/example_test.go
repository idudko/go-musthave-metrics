package pool

import (
	"fmt"
	"sync"
	"testing"
)

// BufferedStruct demonstrates a type that can be pooled.
type BufferedStruct struct {
	Name    string
	Count   int
	Data    []byte
	Items   []string
	Counter int
}

// Reset resets BufferedStruct to its zero values.
func (e *BufferedStruct) Reset() {
	e.Name = ""
	e.Count = 0
	e.Data = e.Data[:0]
	e.Items = e.Items[:0]
	e.Counter = 0
}

func ExamplePool_basicUsage() {
	// Create a new pool for BufferedStruct
	p := New(func() *BufferedStruct {
		return &BufferedStruct{}
	})

	// Get an object from pool
	obj := p.Get()

	// Use the object
	obj.Name = "Test"
	obj.Count = 42
	obj.Data = []byte{1, 2, 3}
	obj.Items = []string{"item1", "item2"}

	fmt.Printf("Before Put: Name=%s, Count=%d\n", obj.Name, obj.Count)

	// Put object back to pool (Reset is called automatically)
	p.Put(obj)

	// Get same object (or a new one if pool was empty)
	obj2 := p.Get()

	fmt.Printf("After Get: Name=%s, Count=%d\n", obj2.Name, obj2.Count)
}

// Output: Before Put: Name=Test, Count=42
// After Get: Name=, Count=0

func ExamplePool_concurrentUsage() {
	// Create a pool with initial capacity
	p := New(func() *BufferedStruct {
		return &BufferedStruct{}
	})

	var wg sync.WaitGroup
	numWorkers := 10

	// Launch multiple goroutines using pool
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Get object from pool
			obj := p.Get()

			// Modify it
			obj.Name = fmt.Sprintf("Worker-%d", id)
			obj.Count = id * 10
			obj.Data = []byte{byte(id)}
			obj.Items = []string{fmt.Sprintf("item-%d", id)}

			// Simulate work
			obj.Counter++

			// Return to pool (will be reset)
			p.Put(obj)
		}(i)
	}

	wg.Wait()

	fmt.Println("All workers completed successfully")
}

// Output: All workers completed successfully

func ExamplePool_customFactory() {
	// Create a pool with custom factory function
	p := New(func() *BufferedStruct {
		// Pre-allocate slices for better performance
		return &BufferedStruct{
			Data:  make([]byte, 0, 1024),
			Items: make([]string, 0, 100),
		}
	})

	obj := p.Get()
	capData := cap(obj.Data)
	capItems := cap(obj.Items)

	fmt.Printf("Data cap: %d, Items cap: %d\n", capData, capItems)

	p.Put(obj)

	// When we get it again, capacity is preserved
	obj2 := p.Get()
	capData2 := cap(obj2.Data)
	capItems2 := cap(obj2.Items)

	fmt.Printf("After reuse - Data cap: %d, Items cap: %d\n", capData2, capItems2)
}

// Output: Data cap: 1024, Items cap: 100
// After reuse - Data cap: 1024, Items cap: 100

func ExamplePool_memoryReuse() {
	// Demonstrate memory reuse benefit
	p := New(func() *BufferedStruct {
		return &BufferedStruct{
			Data:  make([]byte, 0, 1000),
			Items: make([]string, 0, 100),
		}
	})

	// First usage
	obj1 := p.Get()
	obj1.Data = append(obj1.Data, make([]byte, 500)...)
	obj1.Items = append(obj1.Items, "item1", "item2", "item3")

	fmt.Printf("Lengths - Data: %d, Items: %d\n", len(obj1.Data), len(obj1.Items))

	p.Put(obj1)

	// Second usage - memory is reused
	obj2 := p.Get()
	fmt.Printf("After reset - Data: %d, Items: %d\n", len(obj2.Data), len(obj2.Items))

	// Fill it again without reallocation
	obj2.Data = append(obj2.Data, make([]byte, 500)...)
	obj2.Items = append(obj2.Items, "a", "b", "c")

	fmt.Printf("Lengths - Data: %d, Items: %d\n", len(obj2.Data), len(obj2.Items))
}

// Output: Lengths - Data: 500, Items: 3
// After reset - Data: 0, Items: 0
// Lengths - Data: 500, Items: 3

func BenchmarkPoolWithoutPool(b *testing.B) {
	// Benchmark without pool - allocate new objects
	b.ReportAllocs()
	for b.Loop() {
		obj := &BufferedStruct{}
		obj.Name = "test"
		obj.Count = 42
		obj.Data = append(obj.Data, 1, 2, 3)
		obj.Items = append(obj.Items, "a", "b")
		obj.Counter++
		// Use Name and Count to avoid linter warnings
		_ = obj.Name
		obj.Count++
		_ = obj
	}
}

func BenchmarkPoolWithPool(b *testing.B) {
	// Benchmark with pool - reuse objects
	b.ReportAllocs()

	p := New(func() *BufferedStruct {
		return &BufferedStruct{
			Data:  make([]byte, 0, 10),
			Items: make([]string, 0, 10),
		}
	})

	for b.Loop() {
		obj := p.Get()
		obj.Name = "test"
		obj.Count = 42
		obj.Data = append(obj.Data, 1, 2, 3)
		obj.Items = append(obj.Items, "a", "b")
		obj.Counter++
		obj.Count++ // Use Count to avoid linter warning
		p.Put(obj)
	}
}
