package reflection

import "reflect"

var (
	TypeOfError          = reflect.TypeOf((*error)(nil)).Elem()
	TypeOfEmptyInterface = reflect.TypeOf((*interface{})(nil)).Elem()
)
