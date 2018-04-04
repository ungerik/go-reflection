package reflection

import (
	"fmt"
	"reflect"
	"strings"
)

var TypeOfError = reflect.TypeOf((*error)(nil)).Elem()

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

func DerefValue(val interface{}) reflect.Value {
	v := ValueOf(val)
	for v.Kind() == reflect.Ptr {
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
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v, v.Type()
}

// IsNil returns if val is of a type that can be nil and if it is nil.
// Unlike reflect.Value.IsNil() it is safe to call this function for any value and type.
func IsNil(val interface{}) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}
	return false
}

type StructFieldValue struct {
	Field reflect.StructField
	Value reflect.Value
}

// FlatExportedStructFields returns a slice of StructFieldValue of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func FlatExportedStructFields(val interface{}) []StructFieldValue {
	v, t := DerefValueAndType(val)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("FlatExportedStructFields expects struct, pointer to or reflect.Value of a struct argument, but got: %T", val))
	}
	numField := t.NumField()
	fields := make([]StructFieldValue, 0, numField)
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.Anonymous {
			fields = append(fields, FlatExportedStructFields(fieldValue)...)
		} else {
			if fieldType.PkgPath == "" {
				fields = append(fields, StructFieldValue{fieldType, fieldValue})
			}
		}
	}
	return fields
}

// EnumFlatExportedStructFields returns reflect.StructField and reflect.Value of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func EnumFlatExportedStructFields(val interface{}, callback func(reflect.StructField, reflect.Value)) {
	v, t := DerefValueAndType(val)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("EnumFlatExportedStructFields expects struct, pointer to or reflect.Value of a struct argument, but got: %T", val))
	}
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.Anonymous {
			EnumFlatExportedStructFields(fieldValue, callback)
		} else {
			if fieldType.PkgPath == "" {
				callback(fieldType, fieldValue)
			}
		}
	}
}

func exportedFieldName(field reflect.StructField, nameTag string) (name string, valid bool) {
	if field.PkgPath != "" {
		return "", false
	}
	name, ok := field.Tag.Lookup(nameTag)
	if !ok {
		return field.Name, true
	}
	if pos := strings.IndexRune(name, ','); pos != -1 {
		name = name[:pos]
	}
	if name == "-" {
		return "", false
	}
	return name, true
}

type StructFieldValueName struct {
	Field reflect.StructField
	Value reflect.Value
	Name  string
}

// FlatExportedStructFieldValueNames returns a slice of StructFieldValueName of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func FlatExportedStructFieldValueNames(val interface{}, nameTag string) []StructFieldValueName {
	v, t := DerefValueAndType(val)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("FlatExportedStructFieldValueNames expects struct, pointer to or reflect.Value of a struct argument, but got: %T", val))
	}
	numField := t.NumField()
	fields := make([]StructFieldValueName, 0, numField)
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.Anonymous {
			fields = append(fields, FlatExportedStructFieldValueNames(fieldValue, nameTag)...)
		} else {
			if name, valid := exportedFieldName(fieldType, nameTag); valid {
				fields = append(fields, StructFieldValueName{fieldType, fieldValue, name})
			}
		}
	}
	return fields
}

type StructFieldName struct {
	Field reflect.StructField
	Name  string
}

// FlatExportedStructFieldNames returns a slice of StructFieldName of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func FlatExportedStructFieldNames(t reflect.Type, nameTag string) []StructFieldName {
	t = DerefType(t)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("FlatExportedStructFieldNames expects struct, pointer to or reflect.Value of a struct argument, but got: %s", t))
	}
	numField := t.NumField()
	fields := make([]StructFieldName, 0, numField)
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		if field.Anonymous {
			fields = append(fields, FlatExportedStructFieldNames(field.Type, nameTag)...)
		} else {
			if name, valid := exportedFieldName(field, nameTag); valid {
				fields = append(fields, StructFieldName{field, name})
			}
		}
	}
	return fields
}

// ValuesToInterfaces returns a slice of interface{}
// by calling reflect.Value.Interfac() for all values.
func ValuesToInterfaces(values ...reflect.Value) []interface{} {
	s := make([]interface{}, len(values))
	for i := range values {
		s[i] = values[i].Interface()
	}
	return s
}
