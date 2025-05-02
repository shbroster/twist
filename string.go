package twist

import (
	"errors"
	"fmt"
	"reflect"
)

func toString(v interface{}) (string, error) {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		v = reflect.ValueOf(v).Elem().Interface()
	}
	switch val := v.(type) {
	case string:
		return val, nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		bool:
		return fmt.Sprintf("%v", v), nil
	case fmt.Stringer:
		return val.String(), nil

	default:
		return "", errors.New("value cannot be converted to string")
	}
}
