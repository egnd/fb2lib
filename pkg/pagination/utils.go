package pagination

import (
	"fmt"
	"reflect"
)

// toUInt64 convert any numeric value to uint64
func toUInt64(value interface{}) (d uint64, err error) {
	val := reflect.ValueOf(value)

	switch value.(type) {
	case float32, float64:
		d = uint64(val.Float())
	case int, int8, int16, int32, int64:
		d = uint64(val.Int())
	case uint, uint8, uint16, uint32, uint64:
		d = val.Uint()
	default:
		err = fmt.Errorf("toUInt64 need numeric not `%T`", value)
	}

	return
}
