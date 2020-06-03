package portal

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

func nestedValue(ctx context.Context, any interface{}, chainingAttrs []string, cg *cacheGroup, enableCache bool) (interface{}, error) {
	if len(chainingAttrs) == 0 {
		return any, nil
	}

	if any == interface{}(nil) {
		return nil, errors.New("object is nil")
	}

	rv := reflect.ValueOf(any)
	attr := chainingAttrs[0]

	meth, err := findMethod(rv, attr)
	if err != nil {
		if reflect.Indirect(rv).Kind() == reflect.Struct {
			field := reflect.Indirect(rv).FieldByName(attr)
			if field.IsValid() {
				return nestedValue(ctx, field.Interface(), chainingAttrs[1:], nil, false)
			}
		}

		// ignore method not found here.
		// do nothing for mismatched fields.
		logger.Warnf("[portal.nestedValue] %s", err)
		return nil, nil
	}

	var ret interface{}
	if enableCache {
		cacheKey := genCacheKey(ctx, any, any, attr)
		ret, err = invokeWithCache(ctx, rv, meth, attr, cg, cacheKey)
	} else {
		ret, err = invoke(ctx, rv, meth, attr)
	}

	if err != nil {
		return nil, err
	}
	return nestedValue(ctx, ret, chainingAttrs[1:], cg, enableCache)
}

// invokeMethodOfAnyType calls the specified method of a value and return results.
// Note:
// - Context param is optional
// - Method must returns at least one result.
// - Max number of return values is two, and the last one must be of `error` type.
//
// Supported method definitions:
// - `func (f *FooType) Bar(v interface{}) error`
// - `func (f *FooType) Bar(v interface{}) string`
// - `func (f *FooType) Bar(ctx context.Context, v interface{}) error`
// - `func (f *FooType) Bar(ctx context.Context, v interface{}) string`
// - `func (f *FooType) Bar(ctx context.Context, v interface{}) (string, error)`
// - `func (f *FooType) Bar(ctx context.Context, v interface{}) (string, error)`
func invokeMethodOfAnyType(ctx context.Context, any interface{}, name string, args ...interface{}) (interface{}, error) {
	return invokeMethodOfReflectedValue(ctx, reflect.ValueOf(any), name, args...)
}

func invokeMethodOfAnyTypeWithCache(ctx context.Context, any interface{}, name string, cg *cacheGroup, cacheKey *string, args ...interface{}) (interface{}, error) {
	return invokeMethodOfReflectedValueWithCache(ctx, reflect.ValueOf(any), name, cg, cacheKey, args...)
}

func invokeMethodOfReflectedValue(ctx context.Context, any reflect.Value, name string, args ...interface{}) (interface{}, error) {
	method, err := findMethod(any, name)
	if err != nil {
		return nil, err
	}
	return invoke(ctx, any, method, name, args...)
}

func invokeMethodOfReflectedValueWithCache(ctx context.Context, any reflect.Value, name string, cg *cacheGroup, cacheKey *string, args ...interface{}) (interface{}, error) {
	method, err := findMethod(any, name)
	if err != nil {
		return nil, err
	}
	return invokeWithCache(ctx, any, method, name, cg, cacheKey, args...)
}

func invoke(ctx context.Context, any reflect.Value, method reflect.Value, methodName string, args ...interface{}) (interface{}, error) {
	methodType := method.Type()
	if shouldWithContext(methodType) {
		args = append([]interface{}{ctx}, args...)
	}

	methodNameRepr := fmt.Sprintf("%s.%s", any.Type().String(), methodName)

	numIn := methodType.NumIn()
	if numIn > len(args) {
		return reflect.ValueOf(nil), fmt.Errorf("method '%s' must has minimum %d params: %d", methodNameRepr, numIn, len(args))
	}
	if numIn != len(args) && !methodType.IsVariadic() {
		return reflect.ValueOf(nil), fmt.Errorf("method '%s' must has %d params: %d", methodNameRepr, numIn, len(args))
	}

	numOut := methodType.NumOut()
	switch numOut {
	case 1:
		// Cases like:
		// func (f *FooType) Bar() error
		// func (f *FooType) Bar() string
	case 2:
		// Cases like:
		// func (f *FooType) Bar() (string, error)
		if !methodType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return reflect.ValueOf(nil), fmt.Errorf("the last return value of method '%s' must be of `error` type", methodNameRepr)
		}
	default:
		return reflect.ValueOf(nil), fmt.Errorf("method '%s' must returns one result with an optional error", methodNameRepr)
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
			return reflect.ValueOf(nil), fmt.Errorf("param[%d] of method '%s' must be %s, not %s", i, methodNameRepr, argType, inType)
		}
	}

	outs := method.Call(in)
	switch len(outs) {
	case 1:
		return outs[0].Interface(), nil
	case 2:
		err := outs[1].Interface()
		if err != nil {
			return nil, errors.WithStack(err.(error))
		}
		return outs[0].Interface(), nil
	default:
		return nil, errors.Errorf("unexpected results returned by method '%s'", methodNameRepr)
	}
}

func invokeWithCache(ctx context.Context, any reflect.Value, method reflect.Value, methodName string, cg *cacheGroup, cacheKey *string, args ...interface{}) (interface{}, error) {
	if !cg.Valid() || cacheKey == nil {
		ret, err := invoke(ctx, any, method, methodName, args...)
		return ret, errors.WithStack(err)
	}

	cg.mu.Lock()
	ret, err := cg.cache.Get(ctx, *cacheKey)
	if err == nil {
		cg.mu.Unlock()
		return ret, nil
	}

	if cg.m == nil {
		cg.m = make(map[interface{}]*call)
	}
	if c, ok := cg.m[*cacheKey]; ok {
		cg.mu.Unlock()
		c.wg.Wait()
		return c.val, errors.WithStack(c.err)
	}

	c := new(call)
	c.wg.Add(1)
	cg.m[cacheKey] = c
	cg.mu.Unlock()

	c.val, c.err = invoke(ctx, any, method, methodName, args...)
	c.wg.Done()

	cg.mu.Lock()
	delete(cg.m, cacheKey)
	cg.mu.Unlock()

	if c.err == nil {
		cg.cache.Set(ctx, *cacheKey, ret)
		return ret, nil
	}
	return c.val, errors.WithStack(c.err)
}

func findMethod(any reflect.Value, name string) (reflect.Value, error) {
	var vptr = any
	if any.Kind() != reflect.Ptr {
		vptr = reflect.New(any.Type())
		vptr.Elem().Set(any)
	}

	method := vptr.MethodByName(name)
	if method.IsValid() {
		return method, nil
	} else {
		return reflect.Value{}, fmt.Errorf("method '%s' not found in '%s'", name, any.Type().String())
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
