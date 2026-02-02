//go:generate go run ../../cmd/reset/main.go

package resetexample

// generate:reset
type ResetableStruct struct {
	i      int
	str    string
	b      bool
	strP   *string
	s      []int
	m      map[string]string
	child  *ResetableStruct
	slice  []ResetableStruct
	mapped map[int]*ResetableStruct
}

// generate:reset
type NestedStruct struct {
	name string
	age  int
}

// generate:reset
type ComplexStruct struct {
	primitive      int
	stringVal      string
	boolVal        bool
	floatVal       float64
	ptrInt         *int
	ptrString      *string
	sliceInt       []int
	sliceString    []string
	mapIntString   map[int]string
	mapStringInt   map[string]int
	nested         *NestedStruct
	sliceNested    []*NestedStruct
	sliceOfStructs []ComplexStruct
}
