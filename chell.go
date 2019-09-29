package portal

import (
	"context"
	"errors"
	"reflect"
	"sync"
)

var (
	ConcurrentDumpingPoolSize = 10
)

type Chell struct {
	schema interface{} //nolint

	onlyFieldNames     []string
	excludedFieldNames []string
}

func New() *Chell {
	chell := &Chell{
		onlyFieldNames:     make([]string, 0),
		excludedFieldNames: make([]string, 0),
	}
	return chell
}

func Dump(dst interface{}, src interface{}) error {
	return New().Dump(dst, src)
}

func DumpWithContext(ctx context.Context, dst interface{}, src interface{}) error {
	return New().DumpWithContext(ctx, dst, src)
}

func (c *Chell) Dump(dst, src interface{}) error {
	return c.DumpWithContext(context.TODO(), dst, src)
}

func (c *Chell) DumpWithContext(ctx context.Context, dst, src interface{}) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr {
		return errors.New("dst must be a pointer")
	}

	if reflect.Indirect(rv).Kind() == reflect.Slice {
		return c.dumpMany(ctx, dst, src)
	} else {
		toSchema := NewSchema(dst)
		toSchema.SetOnlyFields(c.onlyFieldNames...)
		toSchema.SetExcludeFields(c.excludedFieldNames...)
		return c.dump(ctx, toSchema, src)
	}
}

func (c *Chell) dump(ctx context.Context, dst *Schema, src interface{}) error {
	if err := c.dumpSyncFields(ctx, dst, src); err != nil {
		return err
	}
	return c.dumpAsyncFields(ctx, dst, src)
}

func (c *Chell) dumpSyncFields(ctx context.Context, dst *Schema, src interface{}) error {
	logger.Debugln("[portal.chell] dump sync fields")
	for _, field := range dst.SyncFields() {
		logger.Debugf("[portal.chell] processing sync field '%s'", field)
		val, err := dst.FieldValueFromData(ctx, field, src)
		if err != nil {
			return err
		}
		logger.Debugf("[portal.chell] sync field '%s' got value '%v'", field, val)
		err = c.dumpField(ctx, field, val)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Chell) dumpAsyncFields(ctx context.Context, dst *Schema, src interface{}) error {
	logger.Debugln("[portal.chell] dump async fields")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type Item struct {
		field *Field
		data  interface{}
		err   error
	}

	var wg sync.WaitGroup
	items := make(chan *Item, MinInt(len(dst.AsyncFields()), ConcurrentDumpingPoolSize))

	for _, field := range dst.AsyncFields() {
		wg.Add(1)
		go func(f *Field) {
			defer wg.Done()
			logger.Debugf("[portal.chell] processing sync field '%s'", f)
			val, err := dst.FieldValueFromData(ctx, f, src)
			logger.Debugf("[portal.chell] sync field '%s' got value '%v'", f, val)
			items <- &Item{f, val, err}
		}(field)
	}

	go func() {
		wg.Wait()
		close(items)
	}()

	for item := range items {
		if item.err != nil {
			cancel()
			return item.err
		}

		err := c.dumpField(ctx, item.field, item.data)
		if err != nil {
			cancel()
			return err
		}
	}

	return nil
}

func (c *Chell) dumpField(ctx context.Context, field *Field, src interface{}) error {
	if IsNil(src) {
		logger.Warnf("[portal.chell] cannot get value for field %s, current input value is %v", field, src)
		return nil
	}

	if AreIdenticalType(src, field.Field.Value()) {
		return field.SetValue(src)
	}
	if !field.IsNested() {
		logger.Debugf("[portal.chell] dump normal field %s with value %v", field, src)
		return field.SetValue(src)
	} else {
		if field.Many() {
			logger.Debugf("[portal.chell] dump nested slice field %s with value %v", field, src)
			return c.dumpFieldNestedMany(ctx, field, src)
		} else {
			logger.Debugf("[portal.chell] dump nested field %s with value %v", field, src)
			return c.dumpFieldNestedOne(ctx, field, src)
		}
	}
}

func (c *Chell) dumpFieldNestedOne(ctx context.Context, field *Field, src interface{}) error {
	val := reflect.New(IndirectStructTypeP(reflect.TypeOf(field.Value())))
	toNestedSchema := NewSchema(val.Interface())
	toNestedSchema.SetOnlyFields(field.NestedOnlyNames()...)
	toNestedSchema.SetExcludeFields(field.NestedExcludeNames()...)

	err := c.dump(ctx, toNestedSchema, src)
	if err != nil {
		return err
	}
	switch field.Kind() {
	case reflect.Ptr:
		return field.SetValue(val.Interface())
	case reflect.Struct:
		return field.SetValue(val.Elem().Interface())
	default:
		panic("invalid nested schema")
	}
}

func (c *Chell) dumpFieldNestedMany(ctx context.Context, field *Field, src interface{}) error {
	typ := reflect.TypeOf(field.Value())
	nestedSchemaSlice := reflect.New(typ)

	cpy := c.Only(field.NestedOnlyNames()...).Exclude(field.NestedExcludeNames()...)
	err := cpy.dumpMany(ctx, nestedSchemaSlice.Interface(), src)
	if err != nil {
		return err
	}

	switch typ.Kind() {
	case reflect.Ptr:
		err = field.SetValue(nestedSchemaSlice.Interface())
	case reflect.Slice:
		err = field.SetValue(nestedSchemaSlice.Elem().Interface())
	default:
		panic("invalid nested schema")
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Chell) dumpMany(ctx context.Context, dst, src interface{}) error {
	reflectedData := reflect.ValueOf(src)
	if reflectedData.Kind() != reflect.Slice {
		panic("input src must be a slice")
	}

	schemaSlice := reflect.Indirect(reflect.ValueOf(dst))
	schemaSlice.Set(reflect.MakeSlice(schemaSlice.Type(), reflectedData.Len(), reflectedData.Cap()))

	schemaType := IndirectStructTypeP(schemaSlice.Type())

	var wg sync.WaitGroup

	type Item struct {
		index     int
		schemaPtr reflect.Value
		err       error
	}

	items := make(chan *Item, MinInt(reflectedData.Len(), ConcurrentDumpingPoolSize))

	for i := 0; i < reflectedData.Len(); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			schemaPtr := reflect.New(schemaType)
			toSchema := NewSchema(schemaPtr.Interface())
			toSchema.SetOnlyFields(c.onlyFieldNames...)
			toSchema.SetExcludeFields(c.excludedFieldNames...)
			val := reflectedData.Index(index).Interface()
			err := c.dump(ctx, toSchema, val)
			items <- &Item{
				index:     index,
				schemaPtr: schemaPtr,
				err:       err,
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(items)
	}()

	for item := range items {
		if item.err != nil {
			return item.err
		}

		elem := schemaSlice.Index(item.index)
		switch elem.Kind() {
		case reflect.Struct:
			elem.Set(reflect.Indirect(item.schemaPtr))
		case reflect.Ptr:
			elem.Set(item.schemaPtr)
		}
	}
	return nil
}

func (c *Chell) Only(fields ...string) *Chell {
	cpy := c.clone()
	cpy.onlyFieldNames = fields
	return cpy
}

func (c *Chell) Exclude(fields ...string) *Chell {
	cpy := c.clone()
	cpy.excludedFieldNames = fields
	return cpy
}

func (c *Chell) clone() *Chell {
	cpy := &Chell{
		onlyFieldNames:     c.onlyFieldNames,
		excludedFieldNames: c.excludedFieldNames,
	}
	return cpy
}
