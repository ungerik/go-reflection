package reflection

import (
	"fmt"
	"reflect"
	"strings"
)

// IsZero returns true if the underlying value of v is the zero (default) value of its type,
// or if v itself is nil.
//
// This uses reflect.DeepEqual to compare the value with its zero value,
// so it works correctly for structs, slices, maps, and other composite types.
//
// Example:
//
//	fmt.Println(reflection.IsZero(0))        // true
//	fmt.Println(reflection.IsZero(""))       // true
//	fmt.Println(reflection.IsZero(nil))      // true
//	fmt.Println(reflection.IsZero("hello"))  // false
//	fmt.Println(reflection.IsZero([]int{}))  // true (empty slice is zero)
func IsZero(v any) bool {
	return v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}

// ZeroValueExportedStructFieldNames returns the names of exported struct fields that have zero (default) values.
//
// Parameters:
//   - st: The struct value to examine (can be a struct, pointer to struct, or reflect.Value)
//   - namePrefix: A prefix to add to all returned field names
//   - nameTag: The struct tag key to use for field names (e.g., "json"). If empty or not found, uses Go field name
//   - namesToValidate: Optional list of specific field names to check. If empty, checks all fields
//
// Behavior:
//   - Anonymous embedded structs are flattened
//   - Named sub-structs are checked recursively with their name as prefix (e.g., "Address.Street")
//   - Zero elements in arrays/slices are reported with index notation (e.g., "Items[1]")
//   - Struct tag values can include comma-separated options; only the part before the comma is used
//   - Fields with tag value "-" are ignored
//
// Example:
//
//	type Form struct {
//	    Name  string   `json:"name"`
//	    Email string   `json:"email"`
//	    Age   int      `json:"age"`
//	    Tags  []string `json:"tags"`
//	}
//
//	form := Form{Name: "John", Tags: []string{"a", "", "c"}}
//	zeros := reflection.ZeroValueExportedStructFieldNames(form, "", "json")
//	// zeros: ["email", "age", "tags[1]"]
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

// FieldError represents a validation error for a specific struct field.
// It combines the field name with its validation error.
type FieldError struct {
	FieldName  string // Name of the field that failed validation
	FieldError error  // The validation error for this field
}

// Error implements the error interface, formatting the error as "FieldName: error message".
func (f FieldError) Error() string {
	return f.FieldName + ": " + f.FieldError.Error()
}

// ValidateStructFields validates all exported fields of a struct using a custom validation function.
//
// The validation function is called three times for each field (if applicable):
//  1. With the field value itself
//  2. With the field value's address (if addressable)
//  3. With the dereferenced value (if the field is a pointer and not nil)
//
// This allows the validation function to work with values, pointers, and interface implementations.
//
// Parameters:
//   - validateFunc: Function that validates a value and returns an error if invalid
//   - st: The struct to validate (can be a struct, pointer to struct, or reflect.Value)
//   - namePrefix: A prefix to add to all field names in errors
//   - nameTag: The struct tag key to use for field names (e.g., "json"). If empty, uses Go field name
//   - namesToValidate: Optional list of specific field names to validate. If empty, validates all fields
//
// Behavior:
//   - Anonymous embedded structs are flattened
//   - Named sub-structs are validated recursively
//   - Array and slice elements are validated individually
//   - Returns a slice of FieldError for all fields that failed validation
//
// Example:
//
//	func validateNotEmpty(val any) error {
//	    if s, ok := val.(string); ok && strings.TrimSpace(s) == "" {
//	        return errors.New("cannot be empty")
//	    }
//	    return nil
//	}
//
//	type User struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	user := User{Name: "", Email: "test@example.com"}
//	errors := reflection.ValidateStructFields(validateNotEmpty, user, "", "json")
//	// errors: [FieldError{FieldName: "name", FieldError: errors.New("cannot be empty")}]
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
