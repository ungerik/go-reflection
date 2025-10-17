// Package reflection provides utilities that extend Go's reflect package
// with practical helper functions for common reflection tasks.
//
// The package offers:
//   - Safe pointer dereferencing and nil checking
//   - Struct field manipulation with support for anonymous embedded fields
//   - Field validation with custom validation functions
//   - Zero value detection for structs and fields
//   - Value conversion utilities
//
// Most functions that work with structs support flattening of anonymous
// embedded fields, making them appear as top-level fields. This is useful
// for ORM-like operations, serialization, and validation.
//
// Example usage:
//
//	type User struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	user := User{Name: "Alice", Email: "alice@example.com"}
//
//	// Get field names from struct tags
//	names := reflection.FlatStructFieldTagsOrNames(reflect.TypeOf(user), "json")
//	// names: ["name", "email"]
//
//	// Iterate over fields
//	for field, value := range reflection.FlatExportedStructFieldsIter(user) {
//	    fmt.Printf("%s = %v\n", field.Name, value.Interface())
//	}
package reflection

import "reflect"

var (
	// TypeOfError is the reflect.Type of the error interface.
	// Useful for type comparisons when working with functions that return errors.
	//
	// Example:
	//
	//	func returnsError(fn reflect.Type) bool {
	//	    return fn.NumOut() > 0 && fn.Out(0) == reflection.TypeOfError
	//	}
	TypeOfError = reflect.TypeOf((*error)(nil)).Elem()

	// TypeOfEmptyInterface is the reflect.Type of the empty interface (any).
	// Useful for type comparisons when working with generic interfaces.
	TypeOfEmptyInterface = reflect.TypeOf((*interface{})(nil)).Elem()
)
