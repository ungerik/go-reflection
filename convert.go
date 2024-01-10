package reflection

import (
	"reflect"
)

// ValuesToInterfaces returns a slice of interface{}
// by calling reflect.Value.Interfac() for all values.
func ValuesToInterfaces(values ...reflect.Value) []any {
	s := make([]any, len(values))
	for i := range values {
		s[i] = values[i].Interface()
	}
	return s
}
