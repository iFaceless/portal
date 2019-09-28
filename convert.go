package portal

import (
	"reflect"

	"time"

	"fmt"

	"github.com/spf13/cast"
)

// Convert converts value from src type to target type
// Conditions should be considered carefully:
// - from is value type, to is a pointer type
// - from is pointer type，to is value type
// - from is value type, to is value type
// - from and to are all pointer type
func Convert(from interface{}, to interface{}) (out interface{}, err error) {
	if isEmptyValue(from) {
		return nil, fmt.Errorf("empty input value: %s", from)
	}

	v := from
	iv := reflect.ValueOf(from)
	if iv.Type().Kind() == reflect.Ptr {
		v = iv.Elem().Interface()
	}

	switch to.(type) {
	case int:
		out, err = cast.ToIntE(v)
	case *int:
		out, err = toIntPtrE(v)
	case int64:
		out, err = cast.ToInt64E(v)
	case *int64:
		out, err = toInt64PtrE(v)
	case int32:
		out, err = cast.ToInt32E(v)
	case *int32:
		out, err = toInt32PtrE(v)
	case int16:
		out, err = cast.ToInt16E(v)
	case *int16:
		out, err = toInt16PtrE(v)
	case int8:
		out, err = cast.ToInt8E(v)
	case *int8:
		out, err = toInt8PtrE(v)
	case uint:
		out, err = cast.ToUintE(v)
	case *uint:
		out, err = toUintPtrE(v)
	case uint64:
		out, err = cast.ToUint64E(v)
	case *uint64:
		out, err = toUint64PtrE(v)
	case uint32:
		out, err = cast.ToUint32E(v)
	case *uint32:
		out, err = toUint32PtrE(v)
	case uint16:
		out, err = cast.ToUint16E(v)
	case *uint16:
		out, err = toUint16PtrE(v)
	case uint8:
		out, err = cast.ToUint8E(v)
	case *uint8:
		out, err = toUint8PtrE(v)
	case string:
		out, err = cast.ToStringE(v)
	case *string:
		out, err = toStringPtrE(v)
	case time.Time:
		out, err = cast.ToTimeE(v)
	case *time.Time:
		out, err = toTimePtrE(v)
	case time.Duration:
		out, err = cast.ToDurationE(v)
	case *time.Duration:
		out, err = toDurationPtrE(v)
	case bool:
		out, err = cast.ToBoolE(v)
	case *bool:
		out, err = toBoolPtrE(v)
	case float32:
		out, err = cast.ToFloat32E(v)
	case *float32:
		out, err = toFloat32PtrE(v)
	case float64:
		out, err = cast.ToFloat64E(v)
	case *float64:
		out, err = toFloat64PtrE(v)
	case map[string]string:
		out, err = cast.ToStringMapStringE(v)
	case *map[string]string:
		out, err = toStringMapStringPtrE(v)
	case map[string][]string:
		out, err = cast.ToStringMapStringSliceE(v)
	case *map[string][]string:
		out, err = toStringMapStringSlicePtrE(v)
	case map[string]bool:
		out, err = cast.ToStringMapBoolE(v)
	case *map[string]bool:
		out, err = toStringMapBoolPtrE(v)
	case map[string]interface{}:
		out, err = cast.ToStringMapE(v)
	case *map[string]interface{}:
		out, err = toStringMapPtrE(v)
	case []interface{}:
		out, err = cast.ToSliceE(v)
	case *[]interface{}:
		out, err = toSlicePtrE(v)
	case []bool:
		out, err = cast.ToBoolSliceE(v)
	case *[]bool:
		out, err = toBoolSlicePtrE(v)
	case []string:
		out, err = cast.ToStringSliceE(v)
	case *[]string:
		out, err = toStringSlicePtrE(v)
	case []int:
		out, err = cast.ToIntSliceE(v)
	case *[]int:
		out, err = toIntSlicePtrE(v)
	case []time.Duration:
		out, err = cast.ToDurationSliceE(v)
	case *[]time.Duration:
		out, err = toDurationSlicePtrE(v)
	default:
		return convertUsingReflect(to, from)
	}

	// die trying...
	if err != nil {
		out, err = convertUsingReflect(to, from)
	}

	return
}

func isEmptyValue(in interface{}) bool {
	v := reflect.ValueOf(in)
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	value := reflect.ValueOf(in)
	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

func convertUsingReflect(to interface{}, from interface{}) (interface{}, error) {
	expectedType := reflect.TypeOf(to)
	value := reflect.ValueOf(from)
	if value.Type().ConvertibleTo(expectedType) {
		return value.Convert(expectedType).Interface(), nil
	}
	return nil, fmt.Errorf("failed to convert from type '%s' to '%s'", value.Type().Name(), expectedType.Name())
}

func toIntPtrE(v interface{}) (*int, error) {
	cv, err := cast.ToIntE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toInt64PtrE(v interface{}) (*int64, error) {
	cv, err := cast.ToInt64E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toInt32PtrE(v interface{}) (*int32, error) {
	cv, err := cast.ToInt32E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toInt16PtrE(v interface{}) (*int16, error) {
	cv, err := cast.ToInt16E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toInt8PtrE(v interface{}) (*int8, error) {
	cv, err := cast.ToInt8E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toUintPtrE(v interface{}) (*uint, error) {
	cv, err := cast.ToUintE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toUint64PtrE(v interface{}) (*uint64, error) {
	cv, err := cast.ToUint64E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toUint32PtrE(v interface{}) (*uint32, error) {
	cv, err := cast.ToUint32E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toUint16PtrE(v interface{}) (*uint16, error) {
	cv, err := cast.ToUint16E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toUint8PtrE(v interface{}) (*uint8, error) {
	cv, err := cast.ToUint8E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toStringPtrE(v interface{}) (*string, error) {
	cv, err := cast.ToStringE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toTimePtrE(v interface{}) (*time.Time, error) {
	cv, err := cast.ToTimeE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toDurationPtrE(v interface{}) (*time.Duration, error) {
	cv, err := cast.ToDurationE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toBoolPtrE(v interface{}) (*bool, error) {
	cv, err := cast.ToBoolE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toFloat32PtrE(v interface{}) (*float32, error) {
	cv, err := cast.ToFloat32E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toFloat64PtrE(v interface{}) (*float64, error) {
	cv, err := cast.ToFloat64E(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toStringMapStringPtrE(v interface{}) (*map[string]string, error) {
	cv, err := cast.ToStringMapStringE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toStringMapStringSlicePtrE(v interface{}) (*map[string][]string, error) {
	cv, err := cast.ToStringMapStringSliceE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toStringMapBoolPtrE(v interface{}) (*map[string]bool, error) {
	cv, err := cast.ToStringMapBoolE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toStringMapPtrE(v interface{}) (*map[string]interface{}, error) {
	cv, err := cast.ToStringMapE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toSlicePtrE(v interface{}) (*[]interface{}, error) {
	cv, err := cast.ToSliceE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toBoolSlicePtrE(v interface{}) (*[]bool, error) {
	cv, err := cast.ToBoolSliceE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toStringSlicePtrE(v interface{}) (*[]string, error) {
	cv, err := cast.ToStringSliceE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toIntSlicePtrE(v interface{}) (*[]int, error) {
	cv, err := cast.ToIntSliceE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func toDurationSlicePtrE(v interface{}) (*[]time.Duration, error) {
	cv, err := cast.ToDurationSliceE(v)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}
