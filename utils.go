package portal

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

func nestedValue(ctx context.Context, any interface{}, chainingAttrs []string) (interface{}, error) {
	if len(chainingAttrs) == 0 {
		return any, nil
	}

	if any == interface{}(nil) {
		return nil, errors.New("object is nil")
	}

	objValue := reflect.ValueOf(any)
	if reflect.Indirect(objValue).Kind() != reflect.Struct {
		return nil, errors.New("object must be a struct or a pointer to struct")
	}

	attr := chainingAttrs[0]
	field := reflect.Indirect(objValue).FieldByName(attr)
	if field.IsValid() {
		return nestedValue(ctx, field.Interface(), chainingAttrs[1:])
	} else {
		ret, err := invokeStructMethod(ctx, any, attr)
		if err != nil {
			return nil, err
		}
		return nestedValue(ctx, ret, chainingAttrs[1:])
	}
}

// invokeStructMethod calls the specified method of given struct `any` and return results.
func invokeStructMethod(ctx context.Context, any interface{}, name string, args ...interface{}) (interface{}, error) {
	structValue := reflect.ValueOf(any)
	method, err := findStructMethod(structValue, name)
	if err != nil {
		return nil, err
	}
	methodType := method.Type()
	if shouldWithContext(methodType) {
		args = append([]interface{}{ctx}, args...)
	}

	numIn := methodType.NumIn()
	if numIn > len(args) {
		return reflect.ValueOf(nil), fmt.Errorf("method '%s' must has minimum %d params: %d", name, numIn, len(args))
	}
	if numIn != len(args) && !methodType.IsVariadic() {
		return reflect.ValueOf(nil), fmt.Errorf("method '%s' must has %d params: %d", name, numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return reflect.ValueOf(nil), fmt.Errorf("method '%s', param[%d] must be %s, not %s", name, i, inType, argType)
		}
	}
	return (method.Call(in)[0]).Interface(), nil
}

func findStructMethod(any reflect.Value, name string) (reflect.Value, error) {
	var structPtr = any
	if any.Kind() != reflect.Ptr {
		structPtr = reflect.New(any.Type())
		structPtr.Elem().Set(any)
	}

	method := structPtr.MethodByName(name)
	if method.IsValid() {
		return method, nil
	} else {
		return reflect.Value{}, fmt.Errorf("method '%s' not found in '%s'", name, any.Elem().Type().Name())
	}
}

func shouldWithContext(funcType reflect.Type) bool {
	return funcType.NumIn() > 0 && funcType.In(0).Name() == "Context"
}

// indirectStructTypeP get indirect struct type, panics if failed
func indirectStructTypeP(typ reflect.Type) reflect.Type {
	typ, err := indirectStructTypeE(typ)
	if err != nil {
		panic(fmt.Sprintf("failed to get indirect struct type: %s", err))
	}
	return typ
}

func indirectStructTypeE(typ reflect.Type) (reflect.Type, error) {
	switch typ.Kind() {
	case reflect.Struct:
		return typ, nil
	case reflect.Slice:
		return indirectStructTypeE(typ.Elem())
	case reflect.Ptr:
		return indirectStructTypeE(typ.Elem())
	default:
		return nil, fmt.Errorf("unsupported type '%s'", typ.Name())
	}
}

func structName(v interface{}) string {
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		panic("invalid struct type")
	}
	return typ.Name()
}

func isNil(in interface{}) bool {
	if in == nil {
		return true
	}
	v := reflect.ValueOf(in)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

func convertible(from, to interface{}) bool {
	return reflect.TypeOf(from).ConvertibleTo(reflect.TypeOf(to))
}

// innerStructType gets the inner struct type.
// Cases:
// - ModelStruct
// - &ModelStruct
// - &&ModelStruct
func innerStructType(typ reflect.Type) (reflect.Type, error) {
	switch typ.Kind() {
	case reflect.Struct:
		return typ, nil
	case reflect.Ptr:
		curType := typ
		for ptrLevel := 0; ptrLevel < 2; ptrLevel++ {
			switch curType.Elem().Kind() {
			case reflect.Ptr:
				curType = curType.Elem()
			case reflect.Struct:
				return curType.Elem(), nil
			default:
				return nil, errors.New("failed to get inner struct type")
			}
		}
		return nil, errors.New("pointer level too deep")
	default:
		return nil, errors.New("failed to get inner struct type")
	}
}
