package reflection

import (
	"reflect"
)

// ValuesToInterfaces converts a slice of reflect.Value to a slice of any (interface{})
// by calling reflect.Value.Interface() for each value.
//
// This is useful when you have reflect.Value objects and need to pass them
// to functions that accept any/interface{} parameters.
//
// Example:
//
//	values := []reflect.Value{
//	    reflect.ValueOf(42),
//	    reflect.ValueOf("hello"),
//	    reflect.ValueOf(true),
//	}
//	interfaces := reflection.ValuesToInterfaces(values...)
//	// interfaces: []any{42, "hello", true}
func ValuesToInterfaces(values ...reflect.Value) []any {
	s := make([]any, len(values))
	for i := range values {
		s[i] = values[i].Interface()
	}
	return s
}
