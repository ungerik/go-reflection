package reflection

import (
	"fmt"
	"reflect"
	"strings"
)

// IsZero returns if underlying value of v is the zero (default) value of its type,
// or if v itself is nil.
func IsZero(v any) bool {
	return v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

// ZeroValueExportedStructFieldNames returns the names of exported zero (default) value struct fields.
// If a struct field has a tag with the key nameTag, then the tag value will be used as field name,
// else the Go struct field name will be used.
// All returned names are prefixed with namePrefix.
// Anonymous sub structs will be flattened, named sub structs are checked recursively with their
// name used as prefix delimited with a point.
// Zero array and slice fields will be added with thair name and index formated as "%s[%d]".
func ZeroValueExportedStructFieldNames(st any, namePrefix, nameTag string, namesToValidate ...string) (zeroNames []string) {
	v, t := DerefValueAndType(st)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("%T is not a struct or pointer to a struct", st))
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		fieldName, ext := getFieldName(field, namePrefix, nameTag)
		if ignoreField(namesToValidate, fieldName, ext) {
			continue
		}

		fieldVal := v.Field(i)

		switch kind := fieldVal.Kind(); kind {
		case reflect.Ptr:
			if fieldVal.IsNil() {
				zeroNames = append(zeroNames, fieldName)
				continue
			}
			if fieldVal.Type().Elem().Kind() == reflect.Struct {
				zeroNames = append(zeroNames, ZeroValueExportedStructFieldNames(fieldVal.Interface(), fieldName+".", nameTag, namesToValidate...)...)
				continue
			}

		case reflect.Struct:
			if fieldVal.CanAddr() {
				// Use pointer if possible to avoid copy of struct
				fieldVal = fieldVal.Addr()
			}
			zeroNames = append(zeroNames, ZeroValueExportedStructFieldNames(fieldVal.Interface(), fieldName+".", nameTag, namesToValidate...)...)
			continue

		case reflect.Slice, reflect.Array:
			if kind == reflect.Slice && fieldVal.IsNil() {
				zeroNames = append(zeroNames, fieldName)
				continue
			}
			for j := 0; j < fieldVal.Len(); j++ {
				if IsZero(fieldVal.Index(j).Interface()) {
					zeroNames = append(zeroNames, fmt.Sprintf("%s[%d]", fieldName, j))
				}
			}
			continue

		case reflect.Map:
			if fieldVal.IsNil() {
				zeroNames = append(zeroNames, fieldName)
				continue
			}
			panic("TODO")
		}

		if IsZero(fieldVal.Interface()) {
			zeroNames = append(zeroNames, fieldName)
		}
	}

	return zeroNames
}

func getFieldName(field reflect.StructField, namePrefix, nameTag string) (name string, ext string) {
	name = field.Tag.Get(nameTag)
	if comma := strings.IndexByte(name, ','); comma != -1 {
		name, ext = name[:comma], name[comma+1:]
	}
	if name == "" {
		name = field.Name
	}
	return namePrefix + name, ext
}

func ignoreField(namesToValidate []string, name, _ string) bool {
	if len(namesToValidate) == 0 {
		return strings.Contains(name, "-")
	}
	for _, n := range namesToValidate {
		if n == name {
			return strings.Contains(name, "-")
		}
	}
	return true
}

func validate(validateFunc func(any) error, v reflect.Value) error {
	err := validateFunc(v.Interface())
	if err == nil {
		if v.CanAddr() {
			err = validateFunc(v.Addr().Interface())
			if err == nil {
				if v.Kind() == reflect.Ptr && !v.IsNil() {
					err = validateFunc(v.Elem().Interface())
				}
			}
		}
	}
	return err
}

type FieldError struct {
	FieldName  string
	FieldError error
}

func (f FieldError) Error() string {
	return f.FieldName + ": " + f.FieldError.Error()
}

func ValidateStructFields(validateFunc func(any) error, st any, namePrefix, nameTag string, namesToValidate ...string) (fieldErrors []FieldError) {
	v, t := DerefValueAndType(st)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("%T is not a struct or pointer to a struct", st))
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		fieldName, ext := getFieldName(field, namePrefix, nameTag)
		if ignoreField(namesToValidate, fieldName, ext) {
			continue
		}

		fieldVal := v.Field(i)

		err := validate(validateFunc, fieldVal)
		if err != nil {
			fieldErrors = append(fieldErrors, FieldError{fieldName, err})
		}

		switch kind := fieldVal.Kind(); kind {
		case reflect.Struct:
			if fieldVal.CanAddr() {
				// Use pointer if possible to avoid copy of struct
				fieldVal = fieldVal.Addr()
			}
			fieldErrors = append(fieldErrors, ValidateStructFields(validateFunc, fieldVal.Interface(), fieldName+".", nameTag, namesToValidate...)...)

		case reflect.Slice, reflect.Array:
			for j := 0; j < fieldVal.Len(); j++ {
				err := validate(validateFunc, fieldVal.Index(j))
				if err != nil {
					fieldErrors = append(fieldErrors, FieldError{fmt.Sprintf("%s[%d]", fieldName, j), err})
				}
			}

		case reflect.Map:
			panic("TODO")
		}
	}

	return fieldErrors
}
