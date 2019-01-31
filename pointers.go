package reflection

import "reflect"

// ValueOf differs from reflect.ValueOf in that it returns the argument val
// casted to reflect.Value if val is alread a reflect.Value.
// Else the standard result of reflect.ValueOf(val) will be returned.
func ValueOf(val interface{}) reflect.Value {
	v, ok := val.(reflect.Value)
	if ok {
		return v
	}
	return reflect.ValueOf(val)
}

// DerefValue dereferences val until a non pointer type or nil is found
func DerefValue(val interface{}) reflect.Value {
	v := ValueOf(val)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

func DerefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func DerefValueAndType(val interface{}) (reflect.Value, reflect.Type) {
	v := ValueOf(val)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v, v.Type()
}

// IsNil returns if val is of a type that can be nil and if it is nil.
// Unlike reflect.Value.IsNil() it is safe to call this function for any value and type.
func IsNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}
