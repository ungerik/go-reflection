package reflection

import (
	"fmt"
	"reflect"
)

// IsZero returns if underlying value of v is the zero (default) value of its type,
// or if v itself is nil.
func IsZero(v interface{}) bool {
	return v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

// ZeroValueExportedStructFieldNames returns the names of exported zero (default) value struct fields.
// If a struct field has a tag with the key nameTag, then the tag value will be used as field name,
// else the Go struct field name will be used.
// All returned names are prefixed with namePrefix.
// Anonymous sub structs will be flattened, named sub structs are checked recursively with their
// name used as prefix delimited with a point.
// Zero array and slice fields will be added with thair name and index formated as "%s[%d]".
func ZeroValueExportedStructFieldNames(st interface{}, namePrefix, nameTag string, namesToValidate ...string) (zeroNames []string) {
	v, t := DerefValueAndType(st)
	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("%T is not a struct or pointer to a struct", st))
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath == "" {
			continue // not exported
		}

		fieldVal := v.Field(i)

		fieldName := field.Tag.Get(nameTag)
		if fieldName == "" {
			fieldName = field.Name
		}
		fieldName = namePrefix + fieldName
		if ingoreName(fieldName) {
			continue
		}

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
			zeroNames = append(zeroNames, ZeroValueExportedStructFieldNames(fieldVal.Addr().Interface(), fieldName+".", nameTag, namesToValidate...)...)
			continue

		case reflect.Slice, reflect.Array:
			if kind == reflect.Slice && fieldVal.IsNil() {
				zeroNames = append(zeroNames, fieldName)
				continue
			}
			for j := 0; j < fieldVal.Len(); j++ {
				if IsZero(fieldVal.Index(j)) {
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

func ingoreName(name string, namesToValidate ...string) bool {
	if len(namesToValidate) == 0 {
		return false
	}
	for _, n := range namesToValidate {
		if n == name {
			return false
		}
	}
	return true
}
