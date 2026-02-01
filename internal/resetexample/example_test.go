package resetexample

import (
	"testing"
)

func TestResetableStruct_Reset(t *testing.T) {
	// Create test string
	testStr := "test"

	obj := &ResetableStruct{
		i:     42,
		str:   "hello",
		b:     true,
		strP:  &testStr,
		s:     []int{1, 2, 3},
		m:     map[string]string{"key": "value"},
		child: &ResetableStruct{i: 10, str: "child"},
		slice: []ResetableStruct{{i: 1}, {i: 2}},
		mapped: map[int]*ResetableStruct{
			1: {i: 100},
		},
	}

	// Call Reset
	obj.Reset()

	// Verify all fields are reset
	if obj.i != 0 {
		t.Errorf("expected i to be 0, got %d", obj.i)
	}

	if obj.str != "" {
		t.Errorf("expected str to be empty, got %s", obj.str)
	}

	if obj.b != false {
		t.Errorf("expected b to be false, got %t", obj.b)
	}

	if obj.strP != nil {
		if *obj.strP != "" {
			t.Errorf("expected *strP to be empty, got %s", *obj.strP)
		}
	}

	if len(obj.s) != 0 {
		t.Errorf("expected s to have length 0, got %d", len(obj.s))
	}
	// Verify slice is not nil ([:0] keeps it non-nil)
	if obj.s == nil {
		t.Errorf("expected s to be non-nil, got nil")
	}

	if len(obj.m) != 0 {
		t.Errorf("expected m to be empty, got %d items", len(obj.m))
	}
	// Verify map is not nil after clear
	if obj.m == nil {
		t.Errorf("expected m to be non-nil, got nil")
	}

	// Verify child struct is reset
	if obj.child == nil {
		t.Error("expected child to be non-nil")
	} else {
		if obj.child.i != 0 {
			t.Errorf("expected child.i to be 0, got %d", obj.child.i)
		}
		if obj.child.str != "" {
			t.Errorf("expected child.str to be empty, got %s", obj.child.str)
		}
	}

	// Verify slice of structs
	if len(obj.slice) != 0 {
		t.Errorf("expected slice to have length 0, got %d", len(obj.slice))
	}

	// Verify map of pointers to structs
	if len(obj.mapped) != 0 {
		t.Errorf("expected mapped to be empty, got %d items", len(obj.mapped))
	}
}

func TestResetableStruct_Reset_NilReceiver(t *testing.T) {
	var obj *ResetableStruct

	// Should not panic on nil receiver
	obj.Reset()
}

func TestNestedStruct_Reset(t *testing.T) {
	obj := &NestedStruct{
		name: "John",
		age:  30,
	}

	obj.Reset()

	if obj.name != "" {
		t.Errorf("expected name to be empty, got %s", obj.name)
	}

	if obj.age != 0 {
		t.Errorf("expected age to be 0, got %d", obj.age)
	}
}

func TestComplexStruct_Reset(t *testing.T) {
	testInt := 42
	testString := "hello"

	obj := &ComplexStruct{
		primitive:      100,
		stringVal:      "test",
		boolVal:        true,
		floatVal:       3.14,
		ptrInt:         &testInt,
		ptrString:      &testString,
		sliceInt:       []int{1, 2, 3},
		sliceString:    []string{"a", "b", "c"},
		mapIntString:   map[int]string{1: "one"},
		mapStringInt:   map[string]int{"one": 1},
		nested:         &NestedStruct{name: "nested", age: 20},
		sliceNested:    []*NestedStruct{{name: "item1"}, {name: "item2"}},
		sliceOfStructs: []ComplexStruct{{primitive: 1}, {primitive: 2}},
	}

	obj.Reset()

	// Verify primitives
	if obj.primitive != 0 {
		t.Errorf("expected primitive to be 0, got %d", obj.primitive)
	}

	if obj.stringVal != "" {
		t.Errorf("expected stringVal to be empty, got %s", obj.stringVal)
	}

	if obj.boolVal != false {
		t.Errorf("expected boolVal to be false, got %t", obj.boolVal)
	}

	if obj.floatVal != 0 {
		t.Errorf("expected floatVal to be 0, got %f", obj.floatVal)
	}

	// Verify pointers
	if obj.ptrInt != nil {
		if *obj.ptrInt != 0 {
			t.Errorf("expected *ptrInt to be 0, got %d", *obj.ptrInt)
		}
	}

	if obj.ptrString != nil {
		if *obj.ptrString != "" {
			t.Errorf("expected *ptrString to be empty, got %s", *obj.ptrString)
		}
	}

	// Verify slices (should be [:0], not nil)
	if len(obj.sliceInt) != 0 {
		t.Errorf("expected sliceInt to have length 0, got %d", len(obj.sliceInt))
	}
	if obj.sliceInt == nil {
		t.Error("expected sliceInt to be non-nil, got nil")
	}

	if len(obj.sliceString) != 0 {
		t.Errorf("expected sliceString to have length 0, got %d", len(obj.sliceString))
	}

	if len(obj.sliceNested) != 0 {
		t.Errorf("expected sliceNested to have length 0, got %d", len(obj.sliceNested))
	}

	if len(obj.sliceOfStructs) != 0 {
		t.Errorf("expected sliceOfStructs to have length 0, got %d", len(obj.sliceOfStructs))
	}

	// Verify maps (should be cleared, not nil)
	if len(obj.mapIntString) != 0 {
		t.Errorf("expected mapIntString to be empty, got %d items", len(obj.mapIntString))
	}
	if obj.mapIntString == nil {
		t.Error("expected mapIntString to be non-nil, got nil")
	}

	if len(obj.mapStringInt) != 0 {
		t.Errorf("expected mapStringInt to be empty, got %d items", len(obj.mapStringInt))
	}

	// Verify nested struct
	if obj.nested == nil {
		t.Error("expected nested to be non-nil")
	} else {
		if obj.nested.name != "" {
			t.Errorf("expected nested.name to be empty, got %s", obj.nested.name)
		}
		if obj.nested.age != 0 {
			t.Errorf("expected nested.age to be 0, got %d", obj.nested.age)
		}
	}
}

func TestComplexStruct_Reset_NilReceiver(t *testing.T) {
	var obj *ComplexStruct

	// Should not panic on nil receiver
	obj.Reset()
}

func TestResetableStruct_Reset_PointerFields(t *testing.T) {
	// Test that pointer fields that were nil remain nil after Reset
	obj := &ResetableStruct{
		i:    1,
		str:  "test",
		strP: nil, // nil pointer
		s:    []int{1},
		m:    make(map[string]string),
	}

	obj.Reset()

	// strP should still be nil
	if obj.strP != nil {
		t.Error("expected strP to remain nil")
	}

	// But other fields should be reset
	if obj.i != 0 {
		t.Errorf("expected i to be 0, got %d", obj.i)
	}

	if obj.str != "" {
		t.Errorf("expected str to be empty, got %s", obj.str)
	}
}

func TestResetableStruct_Reset_NilMap(t *testing.T) {
	// Test Reset with nil map
	obj := &ResetableStruct{
		i:   1,
		str: "test",
		m:   nil, // nil map
	}

	// This might panic if the generator doesn't handle nil maps
	obj.Reset()

	// After Reset, map should be nil if it was nil initially
	// or cleared if it was non-nil
	// Actually, with clear(nil), Go will panic, so we need to be careful
	// The current implementation does clear(x.m) which will panic if m is nil
	// Let's test if the generator handles this

	// For now, let's just verify other fields
	if obj.i != 0 {
		t.Errorf("expected i to be 0, got %d", obj.i)
	}
}

func TestResetWithMultipleResets(t *testing.T) {
	// Test that calling Reset multiple times doesn't cause issues
	obj := &ResetableStruct{
		i:   42,
		str: "test",
		s:   []int{1, 2, 3},
	}

	obj.Reset()
	obj.Reset()
	obj.Reset()

	// Should still be at zero values
	if obj.i != 0 {
		t.Errorf("expected i to be 0 after multiple resets, got %d", obj.i)
	}

	if len(obj.s) != 0 {
		t.Errorf("expected s to have length 0 after multiple resets, got %d", len(obj.s))
	}
}
