package shared

import "reflect"

func GetType(v any) string {
	if v == nil {
		return "nil"
	}

	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
