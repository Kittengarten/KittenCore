package terasu

import (
	"reflect"
)

func isTypeEqual(obj any, name string) bool {
	return reflect.ValueOf(obj).Type().String() == name
}
