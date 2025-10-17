package reflection

import "reflect"

// ValueOf is an enhanced version of reflect.ValueOf that handles reflect.Value arguments.
// If val is already a reflect.Value, it returns val as-is without wrapping.
// Otherwise, it returns reflect.ValueOf(val).
//
// This is useful when writing generic reflection code that might receive
// either regular values or already-reflected values.
//
// Example:
//
//	v := reflect.ValueOf(42)
//	result := reflection.ValueOf(v)
//	// result is the same as v, not reflect.ValueOf(v)
//
//	result2 := reflection.ValueOf(42)
//	// result2 is reflect.ValueOf(42)
func ValueOf(val any) reflect.Value {
	v, ok := val.(reflect.Value)
	if ok {
		return v
	}
	return reflect.ValueOf(val)
}

// DerefValue dereferences val until a non-pointer type or nil pointer is found.
// It repeatedly follows pointers until reaching a concrete value or a nil pointer.
//
// The argument val can be any value or a reflect.Value (handled by ValueOf).
//
// Example:
//
//	x := 42
//	ptr := &x
//	ptrPtr := &ptr
//
//	v := reflection.DerefValue(ptrPtr)
//	fmt.Println(v.Int()) // 42
//
//	var nilPtr *int
//	v2 := reflection.DerefValue(nilPtr)
//	fmt.Println(v2.IsValid()) // false (nil pointer)
func DerefValue(val any) reflect.Value {
	v := ValueOf(val)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

// DerefType dereferences a type until a non-pointer type is found.
// It repeatedly follows pointer types until reaching a concrete type.
//
// Example:
//
//	t := reflect.TypeOf((**int)(nil))
//	concrete := reflection.DerefType(t)
//	fmt.Println(concrete.Kind()) // int
func DerefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// DerefValueAndType dereferences val until a non-pointer value is found
// and returns both the dereferenced value and its type.
//
// This is a convenience function that combines DerefValue with getting the type.
//
// Example:
//
//	x := 42
//	ptr := &x
//	value, typ := reflection.DerefValueAndType(ptr)
//	fmt.Println(value.Int(), typ.Kind()) // 42 int
func DerefValueAndType(val any) (reflect.Value, reflect.Type) {
	v := ValueOf(val)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v, v.Type()
}

// IsNil returns true if v is of a nillable type and is nil.
//
// Unlike reflect.Value.IsNil(), this function is safe to call for any value and type.
// It will not panic for non-nillable types (like int, string, etc.).
//
// Returns true for:
//   - Invalid/zero reflect.Value (result of reflect.ValueOf(nil))
//   - Nil pointers, interfaces, channels, functions, maps, or slices
//
// Returns false for:
//   - Non-nillable types (int, string, struct, etc.)
//   - Non-nil values of nillable types
//
// Example:
//
//	var ptr *int
//	fmt.Println(reflection.IsNil(reflect.ValueOf(ptr))) // true
//
//	var num int
//	fmt.Println(reflection.IsNil(reflect.ValueOf(num))) // false (non-nillable type)
//
//	var v reflect.Value
//	fmt.Println(reflection.IsNil(v)) // true (invalid/zero value)
func IsNil(v reflect.Value) bool {
	if !v.IsValid() {
		return true
	}
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}
