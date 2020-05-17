package crowd

import (
	"reflect"
)

func indirect(o interface{}) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(o))
}
