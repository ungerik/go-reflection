package reflection

import (
	"fmt"
	"iter"
	"reflect"
	"strings"
)

// FlatStructFieldCount returns the number of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
func FlatStructFieldCount(t reflect.Type) int {
	t = DerefType(t)
	count := 0
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		f := t.Field(i)
		if f.Anonymous {
			count += FlatStructFieldCount(f.Type)
		} else {
			count++
		}
	}
	return count
}

// FlatStructFieldNames returns the names of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
func FlatStructFieldNames(t reflect.Type) (names []string) {
	t = DerefType(t)
	numField := t.NumField()
	names = make([]string, 0, numField)
	for i := 0; i < numField; i++ {
		f := t.Field(i)
		if f.Anonymous {
			names = append(names, FlatStructFieldNames(f.Type)...)
		} else {
			names = append(names, f.Name)
		}
	}
	return names
}

// FlatStructFieldTags returns the tag values for a tagKey of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// An empty string is returned for fields that don't have a matching tag.
func FlatStructFieldTags(t reflect.Type, tagKey string) (tagValues []string) {
	t = DerefType(t)
	numField := t.NumField()
	tagValues = make([]string, 0, numField)
	for i := 0; i < numField; i++ {
		f := t.Field(i)
		if f.Anonymous {
			tagValues = append(tagValues, FlatStructFieldNames(f.Type)...)
		} else {
			tagValues = append(tagValues, f.Tag.Get(tagKey))
		}
	}
	return tagValues
}

// FlatStructFieldTagsOrNames returns the tag values for tagKey or the names of the field
// if no tag with tagKey is defined at a struct field.
// Fields are flattened,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
func FlatStructFieldTagsOrNames(t reflect.Type, tagKey string) (tagsOrNames []string) {
	t = DerefType(t)
	numField := t.NumField()
	tagsOrNames = make([]string, 0, numField)
	for i := 0; i < numField; i++ {
		f := t.Field(i)
		if f.Anonymous {
			tagsOrNames = append(tagsOrNames, FlatStructFieldNames(f.Type)...)
		} else {
			tagOrName := f.Tag.Get(tagKey)
			if tagOrName == "" {
				tagOrName = f.Name
			}
			tagsOrNames = append(tagsOrNames, tagOrName)
		}
	}
	return tagsOrNames
}

// FlatStructFieldValues returns the values of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
func FlatStructFieldValues(v reflect.Value) (values []reflect.Value) {
	v = DerefValue(v)
	t := v.Type()
	numField := t.NumField()
	values = make([]reflect.Value, 0, numField)
	for i := 0; i < numField; i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		if ft.Anonymous {
			values = append(values, FlatStructFieldValues(fv)...)
		} else {
			values = append(values, fv)
		}
	}
	return values
}

type StructFieldValue struct {
	Field reflect.StructField
	Value reflect.Value
}

// FlatExportedStructFields returns a slice of StructFieldValue of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func FlatExportedStructFields(val any) []StructFieldValue {
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
		} else if fieldType.IsExported() {
			fields = append(fields, StructFieldValue{fieldType, fieldValue})
		}
	}
	return fields
}

// EnumFlatExportedStructFields returns reflect.StructField and reflect.Value of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func EnumFlatExportedStructFields(val any, callback func(reflect.StructField, reflect.Value)) {
	v, t := DerefValueAndType(val)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("EnumFlatExportedStructFields expects struct, pointer to or reflect.Value of a struct argument, but got: %T", val))
	}
	for i := range t.NumField() {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.Anonymous {
			EnumFlatExportedStructFields(fieldValue, callback)
		} else if fieldType.IsExported() {
			callback(fieldType, fieldValue)
		}
	}
}

// FlatExportedStructFieldsIter returns an iterator over flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument s can be a struct, a pointer to a struct, or a reflect.Value.
func FlatExportedStructFieldsIter(s any) iter.Seq2[reflect.StructField, reflect.Value] {
	v, t := DerefValueAndType(s)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("FlatExportedStructFieldsIter expects struct or pointer to or reflect.Value of a struct argument, but got: %T", s))
	}
	return func(yield func(reflect.StructField, reflect.Value) bool) {
		for i := range t.NumField() {
			field, val := t.Field(i), v.Field(i)
			switch {
			case field.Anonymous:
				for fieldA, valA := range FlatExportedStructFieldsIter(val) {
					if !yield(fieldA, valA) {
						return
					}
				}
			case field.IsExported():
				if !yield(field, val) {
					return
				}
			}
		}
	}
}

// nameTag can be empty
func exportedFieldName(field reflect.StructField, nameTag string) (name string, valid bool) {
	if !field.IsExported() {
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
func FlatExportedStructFieldValueNames(val any, nameTag string) []StructFieldValueName {
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

// FlatExportedStructFieldValueNameMap returns a slice of StructFieldValueName of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument val can be a struct, a pointer to a struct, or a reflect.Value.
func FlatExportedStructFieldValueNameMap(val any, nameTag string) map[string]StructFieldValueName {
	fields := make(map[string]StructFieldValueName)
	flatExportedStructFieldValueNameMap(val, nameTag, fields)
	return fields
}

func flatExportedStructFieldValueNameMap(val any, nameTag string, fields map[string]StructFieldValueName) {
	v, t := DerefValueAndType(val)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("FlatExportedStructFieldValueNameMap expects struct, pointer to or reflect.Value of a struct argument, but got: %T", val))
	}
	numField := t.NumField()
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		fieldValue := v.Field(i)
		if fieldType.Anonymous {
			flatExportedStructFieldValueNameMap(fieldValue, nameTag, fields)
		} else {
			if name, valid := exportedFieldName(fieldType, nameTag); valid {
				fields[name] = StructFieldValueName{fieldType, fieldValue, name}
			}
		}
	}
}

type NamedStructField struct {
	Field reflect.StructField
	Name  string
}

// FlatExportedNamedStructFields returns a slice of NamedStructField of flattened struct fields,
// meaning that the fields of anonoymous embedded fields are flattened
// to the top level of the struct.
// The argument t can be a struct, a pointer to a struct, or a reflect.Value.
func FlatExportedNamedStructFields(t reflect.Type, nameTag string) []NamedStructField {
	t = DerefType(t)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("FlatExportedNamedStructFields expects struct, pointer to or reflect.Value of a struct argument, but got: %s", t))
	}
	numField := t.NumField()
	fields := make([]NamedStructField, 0, numField)
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		if field.Anonymous {
			fields = append(fields, FlatExportedNamedStructFields(field.Type, nameTag)...)
		} else {
			if name, valid := exportedFieldName(field, nameTag); valid {
				fields = append(fields, NamedStructField{field, name})
			}
		}
	}
	return fields
}
