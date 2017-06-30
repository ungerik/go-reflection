package reflection

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	TypeOfError = reflect.TypeOf((*error)(nil)).Elem()
)

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

type StructFieldValue struct {
	reflect.StructField
	Value reflect.Value
}

// FlatStructFields returns a slice of StructFieldValue of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func FlatStructFields(val interface{}) []StructFieldValue {
	v, t := DerefValueAndType(val)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("FlatStructFields expects struct, pointer to or reflect.Value of a struct argument, but got: %T", val))
	}
	numField := t.NumField()
	fields := make([]StructFieldValue, 0, numField)
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.Anonymous {
			fields = append(fields, FlatStructFields(fieldValue)...)
		} else {
			fields = append(fields, StructFieldValue{fieldType, fieldValue})
		}
	}
	return fields
}

// FlatStructFields returns reflect.StructField and reflect.Value of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func EnumFlatStructFields(val interface{}, callback func(reflect.StructField, reflect.Value)) {
	v, t := DerefValueAndType(val)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("EnumFlatStructFields expects struct, pointer to or reflect.Value of a struct argument, but got: %T", val))
	}
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.Anonymous {
			EnumFlatStructFields(fieldValue, callback)
		} else {
			callback(fieldType, fieldValue)
		}
	}
}

type StructFieldValueName struct {
	reflect.StructField
	Value reflect.Value
	Name  string
}

// FlatStructFieldValueNames returns a slice of StructFieldValueName of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func FlatStructFieldValueNames(val interface{}, nameTag string) []StructFieldValueName {
	v, t := DerefValueAndType(val)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("FlatStructFieldValueNames expects struct, pointer to or reflect.Value of a struct argument, but got: %T", val))
	}
	numField := t.NumField()
	fields := make([]StructFieldValueName, 0, numField)
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.Anonymous {
			fields = append(fields, FlatStructFieldValueNames(fieldValue, nameTag)...)
		} else {
			name := fieldType.Tag.Get(nameTag)
			if name == "-" {
				continue
			}
			if name == "" {
				name = fieldType.Name
			} else {
				if pos := strings.IndexRune(name, ','); pos != -1 {
					name = name[:pos]
				}
			}
			fields = append(fields, StructFieldValueName{fieldType, fieldValue, name})
		}
	}
	return fields
}
