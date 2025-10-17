# go-reflection

Utilities extending Go's `reflect` package with practical helper functions for common reflection tasks.

[![Go Reference](https://pkg.go.dev/badge/github.com/ungerik/go-reflection.svg)](https://pkg.go.dev/github.com/ungerik/go-reflection)
[![Go Report Card](https://goreportcard.com/badge/github.com/ungerik/go-reflection)](https://goreportcard.com/report/github.com/ungerik/go-reflection)

## Installation

```bash
go get github.com/ungerik/go-reflection
```

## Features

- **Pointer Handling**: Safe dereferencing and nil checking for any value type
- **Type Utilities**: Common type constants and type dereferencing
- **Struct Field Manipulation**: Flatten anonymous embedded fields, iterate, extract values
- **Validation**: Validate struct fields with custom validation functions
- **Zero Value Detection**: Check for zero values in structs and fields
- **Value Conversion**: Convert reflect.Value slices to interface{} slices

## Table of Contents

- [Quick Start](#quick-start)
- [Pointer Utilities](#pointer-utilities)
- [Type Constants](#type-constants)
- [Struct Field Operations](#struct-field-operations)
  - [Flattening Anonymous Fields](#flattening-anonymous-fields)
  - [Field Iteration](#field-iteration)
  - [Field Names and Tags](#field-names-and-tags)
  - [Field Values](#field-values)
- [Validation](#validation)
- [Zero Value Detection](#zero-value-detection)
- [Value Conversion](#value-conversion)

## Quick Start

```go
package main

import (
    "fmt"
    "reflect"
    "github.com/ungerik/go-reflection"
)

type Address struct {
    Street string
    City   string
}

type Person struct {
    Name    string `json:"name"`
    Age     int    `json:"age"`
    Email   string `json:"email"`
    Address        // Anonymous embedded field
}

func main() {
    person := Person{
        Name:  "Alice",
        Age:   30,
        Email: "alice@example.com",
        Address: Address{
            Street: "123 Main St",
            City:   "Springfield",
        },
    }

    // Get flattened field names (including embedded struct fields)
    fieldNames := reflection.FlatStructFieldNames(reflect.TypeOf(person))
    fmt.Println("Fields:", fieldNames)
    // Output: Fields: [Name Age Email Street City]

    // Get field names from struct tags
    tagNames := reflection.FlatStructFieldTagsOrNames(reflect.TypeOf(person), "json")
    fmt.Println("JSON tags:", tagNames)
    // Output: JSON tags: [name age email Street City]

    // Safe nil checking
    var ptr *Person
    fmt.Println("Is nil:", reflection.IsNil(reflect.ValueOf(ptr)))
    // Output: Is nil: true
}
```

## Pointer Utilities

### ValueOf

`ValueOf` is an enhanced version of `reflect.ValueOf` that handles `reflect.Value` arguments:

```go
// If val is already a reflect.Value, return it as-is
// Otherwise, return reflect.ValueOf(val)
v := reflection.ValueOf(someValue)
```

### DerefValue

Dereference a value until a non-pointer type or nil is reached:

```go
var x = 42
ptr := &x
ptrPtr := &ptr

// Dereferences all pointer levels
v := reflection.DerefValue(ptrPtr)
fmt.Println(v.Int()) // 42
```

### DerefType

Dereference a type until a non-pointer type is reached:

```go
t := reflect.TypeOf((**int)(nil))
dereferenced := reflection.DerefType(t)
fmt.Println(dereferenced.Kind()) // int
```

### IsNil

Safe nil checking for any value type:

```go
// Unlike reflect.Value.IsNil(), this is safe for any value type
var ptr *int
fmt.Println(reflection.IsNil(reflect.ValueOf(ptr))) // true

var num int
fmt.Println(reflection.IsNil(reflect.ValueOf(num))) // false

// Safe for invalid values too
var v reflect.Value
fmt.Println(reflection.IsNil(v)) // true
```

## Type Constants

Pre-defined type constants for common interfaces:

```go
// Type of error interface
errorType := reflection.TypeOfError

// Type of empty interface{}
anyType := reflection.TypeOfEmptyInterface

// Use in type comparisons
func returnsError(t reflect.Type) bool {
    return t.NumOut() > 0 && t.Out(0) == reflection.TypeOfError
}
```

## Struct Field Operations

### Flattening Anonymous Fields

Work with embedded (anonymous) struct fields as if they were top-level fields:

```go
type Base struct {
    ID   int
    Name string
}

type Extended struct {
    Base  // Anonymous field
    Email string
}

// Get count of all fields (including embedded)
count := reflection.FlatStructFieldCount(reflect.TypeOf(Extended{}))
fmt.Println(count) // 3 (ID, Name, Email)

// Get all field names
names := reflection.FlatStructFieldNames(reflect.TypeOf(Extended{}))
fmt.Println(names) // [ID Name Email]
```

### Field Iteration

Multiple ways to iterate over struct fields:

```go
type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

user := User{Name: "Bob", Email: "bob@example.com"}

// Method 1: Get slice of field info
fields := reflection.FlatExportedStructFields(user)
for _, field := range fields {
    fmt.Printf("%s = %v\n", field.Field.Name, field.Value.Interface())
}

// Method 2: Use callback
reflection.EnumFlatExportedStructFields(user, func(field reflect.StructField, value reflect.Value) {
    fmt.Printf("%s = %v\n", field.Name, value.Interface())
})

// Method 3: Use iterator (Go 1.23+)
for field, value := range reflection.FlatExportedStructFieldsIter(user) {
    fmt.Printf("%s = %v\n", field.Name, value.Interface())
}
```

### Field Names and Tags

Extract field names from struct tags:

```go
type Product struct {
    ID    int    `json:"id" db:"product_id"`
    Name  string `json:"name" db:"product_name"`
    Price float64 `json:"price" db:"product_price"`
}

// Get json tags
jsonTags := reflection.FlatStructFieldTags(reflect.TypeOf(Product{}), "json")
fmt.Println(jsonTags) // [id name price]

// Get db tags
dbTags := reflection.FlatStructFieldTags(reflect.TypeOf(Product{}), "db")
fmt.Println(dbTags) // [product_id product_name product_price]

// Get tags or field names as fallback
tagsOrNames := reflection.FlatStructFieldTagsOrNames(reflect.TypeOf(Product{}), "xml")
fmt.Println(tagsOrNames) // [ID Name Price] (uses field names since no xml tags)
```

### Field Values

Extract field values as interfaces:

```go
type Point struct {
    X int
    Y int
}

p := Point{X: 10, Y: 20}

// Get field values
values := reflection.FlatStructFieldValues(reflect.ValueOf(p))

// Convert to interface{} slice
interfaces := reflection.ValuesToInterfaces(values...)
fmt.Println(interfaces) // [10 20]
```

### Field Names with Values

Get structured information about fields:

```go
type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

config := Config{Host: "localhost", Port: 8080}

// Get slice of field info with names
fields := reflection.FlatExportedStructFieldValueNames(config, "json")
for _, field := range fields {
    fmt.Printf("%s: %v\n", field.Name, field.Value.Interface())
}
// Output:
// host: localhost
// port: 8080

// Or get as map
fieldMap := reflection.FlatExportedStructFieldValueNameMap(config, "json")
hostField := fieldMap["host"]
fmt.Println(hostField.Value.String()) // localhost
```

## Validation

Validate struct fields using custom validation functions:

```go
import (
    "errors"
    "strings"
)

type UserInput struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
}

// Custom validation function
func validateField(val any) error {
    switch v := val.(type) {
    case string:
        if strings.TrimSpace(v) == "" {
            return errors.New("cannot be empty")
        }
    case int:
        if v < 0 {
            return errors.New("must be positive")
        }
    }
    return nil
}

func main() {
    input := UserInput{
        Username: "",        // Invalid: empty
        Email:    "a@b.com", // Valid
        Age:      -5,        // Invalid: negative
    }

    // Validate all fields
    fieldErrors := reflection.ValidateStructFields(validateField, input, "", "json")

    for _, ferr := range fieldErrors {
        fmt.Printf("Field %s: %v\n", ferr.FieldName, ferr.FieldError)
    }
    // Output:
    // Field username: cannot be empty
    // Field age: must be positive
}
```

The validation function is called with:
1. The field value
2. The field value's address (if addressable)
3. The dereferenced value (if pointer and not nil)

This allows validation functions to work with values, pointers, and implement interface-based validation.

## Zero Value Detection

Check for zero (default) values in structs:

```go
type Form struct {
    Name     string   `json:"name"`
    Email    string   `json:"email"`
    Age      int      `json:"age"`
    Tags     []string `json:"tags"`
    Optional *string  `json:"optional"`
}

func main() {
    form := Form{
        Name:  "John",
        // Email is zero value (empty string)
        Age:   0,  // zero value
        Tags:  nil, // zero value (nil slice)
        // Optional is nil
    }

    // Check if entire value is zero
    fmt.Println(reflection.IsZero(0))        // true
    fmt.Println(reflection.IsZero(""))       // true
    fmt.Println(reflection.IsZero("hello"))  // false

    // Find zero-value fields in struct
    zeroFields := reflection.ZeroValueExportedStructFieldNames(form, "", "json")
    fmt.Println("Zero fields:", zeroFields)
    // Output: Zero fields: [email age tags optional]
}
```

### Advanced Zero Value Detection

```go
type Nested struct {
    Inner struct {
        Value string `json:"value"`
    } `json:"inner"`
    Items []int `json:"items"`
}

nested := Nested{
    Items: []int{1, 0, 3}, // Has a zero element at index 1
}

zeroFields := reflection.ZeroValueExportedStructFieldNames(nested, "", "json")
fmt.Println(zeroFields)
// Output: [inner.value items[1]]
// Note: Shows nested field and array index with zero value
```

## Value Conversion

Convert `reflect.Value` slices to `interface{}` slices:

```go
values := []reflect.Value{
    reflect.ValueOf(42),
    reflect.ValueOf("hello"),
    reflect.ValueOf(true),
}

// Convert to []interface{}
interfaces := reflection.ValuesToInterfaces(values...)
fmt.Printf("%#v\n", interfaces)
// Output: []interface {}{42, "hello", true}
```

## Advanced Examples

### Custom Struct Mapper

Map struct fields to a map using struct tags:

```go
type User struct {
    ID       int    `db:"user_id"`
    Username string `db:"username"`
    Email    string `db:"email"`
}

func structToMap(s any, tagKey string) map[string]any {
    fields := reflection.FlatExportedStructFieldValueNames(s, tagKey)
    result := make(map[string]any, len(fields))
    for _, field := range fields {
        result[field.Name] = field.Value.Interface()
    }
    return result
}

user := User{ID: 1, Username: "alice", Email: "alice@example.com"}
dbMap := structToMap(user, "db")
fmt.Println(dbMap)
// Output: map[email:alice@example.com user_id:1 username:alice]
```

### Validation with Zero Value Check

Combine validation and zero value detection:

```go
type CreateUserRequest struct {
    Username string `json:"username" required:"true"`
    Email    string `json:"email" required:"true"`
    Age      *int   `json:"age"`  // Optional
}

func validateRequired(req any) []string {
    var missing []string

    fields := reflection.FlatExportedStructFieldValueNames(req, "json")
    for _, field := range fields {
        // Check if field has required tag
        if field.Field.Tag.Get("required") == "true" {
            if reflection.IsNil(field.Value) || reflection.IsZero(field.Value.Interface()) {
                missing = append(missing, field.Name)
            }
        }
    }

    return missing
}

req := CreateUserRequest{
    Username: "", // Missing required field
    // Email missing
    Age: nil, // Optional, so OK
}

missing := validateRequired(req)
fmt.Println("Missing required fields:", missing)
// Output: Missing required fields: [username email]
```

### Dynamic Field Filtering

Filter struct fields based on criteria:

```go
type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Price       float64 `json:"price"`
    InternalRef string  `json:"-"` // Excluded by json tag
}

func getJSONFields(s any) []string {
    var fieldNames []string

    reflection.EnumFlatExportedStructFields(s, func(field reflect.StructField, value reflect.Value) {
        jsonTag := field.Tag.Get("json")
        if jsonTag != "" && jsonTag != "-" {
            // Extract field name from tag (before comma)
            if comma := strings.IndexByte(jsonTag, ','); comma != -1 {
                jsonTag = jsonTag[:comma]
            }
            fieldNames = append(fieldNames, jsonTag)
        }
    })

    return fieldNames
}

product := Product{ID: 1, Name: "Widget", Price: 9.99, InternalRef: "WID-001"}
jsonFields := getJSONFields(product)
fmt.Println(jsonFields) // [id name price]
```

## Best Practices

### 1. Prefer Iteration Methods Based on Use Case

```go
// ✅ Good: Use iterator for large structs (memory efficient)
for field, value := range reflection.FlatExportedStructFieldsIter(largeStruct) {
    process(field, value)
}

// ✅ Good: Use slice when you need to modify or reorder
fields := reflection.FlatExportedStructFields(s)
sort.Slice(fields, func(i, j int) bool {
    return fields[i].Field.Name < fields[j].Field.Name
})

// ✅ Good: Use callback when you just need to process each field
reflection.EnumFlatExportedStructFields(s, processField)
```

### 2. Handle Nil Safely

```go
// ✅ Good: Use IsNil for safe nil checking
if reflection.IsNil(v) {
    // Handle nil case
}

// ❌ Avoid: Direct IsNil() call can panic on non-nillable types
if v.IsNil() { // Panics if v is an int, string, etc.
    // ...
}
```

### 3. Cache Type Information

```go
// ✅ Good: Cache type information for repeated operations
type StructInfo struct {
    fieldNames []string
    fieldTags  map[string]string
}

func getStructInfo(t reflect.Type) StructInfo {
    // Cache this result
    return StructInfo{
        fieldNames: reflection.FlatStructFieldNames(t),
        fieldTags:  makeTagMap(t),
    }
}
```

### 4. Use Struct Tags for Metadata

```go
// ✅ Good: Use struct tags to drive reflection behavior
type Model struct {
    ID   int    `db:"id" validate:"required"`
    Name string `db:"name" validate:"required,min=3"`
}

// Then use tags in your reflection code
fields := reflection.FlatExportedStructFieldValueNames(model, "db")
```

## API Reference

See [pkg.go.dev](https://pkg.go.dev/github.com/ungerik/go-reflection) for complete API documentation.

### Core Functions

- `ValueOf(any) reflect.Value` - Enhanced ValueOf that handles reflect.Value arguments
- `DerefValue(any) reflect.Value` - Dereference until non-pointer or nil
- `DerefType(reflect.Type) reflect.Type` - Dereference type until non-pointer
- `IsNil(reflect.Value) bool` - Safe nil checking for any value type
- `IsZero(any) bool` - Check if value is zero value

### Struct Field Functions

- `FlatStructFieldCount(reflect.Type) int` - Count of flattened fields
- `FlatStructFieldNames(reflect.Type) []string` - Names of flattened fields
- `FlatStructFieldTags(reflect.Type, string) []string` - Tag values of flattened fields
- `FlatStructFieldTagsOrNames(reflect.Type, string) []string` - Tags or names as fallback
- `FlatStructFieldValues(reflect.Value) []reflect.Value` - Values of flattened fields
- `FlatExportedStructFields(any) []StructFieldValue` - Field info with values
- `FlatExportedStructFieldsIter(any) iter.Seq2[...]` - Iterator over fields (Go 1.23+)
- `FlatExportedStructFieldValueNames(any, string) []StructFieldValueName` - Fields with tag names
- `FlatExportedStructFieldValueNameMap(any, string) map[string]StructFieldValueName` - Field map by name

### Validation Functions

- `ValidateStructFields(func(any) error, any, string, string, ...string) []FieldError` - Validate fields
- `ZeroValueExportedStructFieldNames(any, string, string, ...string) []string` - Find zero-value fields

### Utility Functions

- `ValuesToInterfaces(...reflect.Value) []any` - Convert Values to interfaces

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](LICENSE) file for details.

## Related Projects

- [go-httpx](https://github.com/ungerik/go-httpx) - HTTP utilities
- [go-fs](https://github.com/ungerik/go-fs) - File system abstraction
- [go-dry](https://github.com/ungerik/go-dry) - DRY utilities for Go
