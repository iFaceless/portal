package portal

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

var (
	ConcurrentDumpingPoolSize = 10
)

type Chell struct {
	schema interface{}

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

func (c *Chell) Dump(ctx context.Context, src interface{}, dest interface{}) error {
	toSchema := NewSchema(dest)
	toSchema.SetOnlyFields(c.onlyFieldNames...)
	toSchema.SetExcludeFields(c.excludedFieldNames...)
	return c.dump(ctx, src, toSchema)
}

func (c *Chell) dump(ctx context.Context, src interface{}, dest *Schema) error {
	if err := c.dumpSyncFields(ctx, src, dest); err != nil {
		return err
	}
	return c.dumpAsyncFields(ctx, src, dest)
}

func (c *Chell) dumpSyncFields(ctx context.Context, src interface{}, dest *Schema) error {
	logger.Debugln("[portal.chell] dump sync fields")
	for _, field := range dest.SyncFields() {
		logger.Debugf("[portal.chell] processing sync field '%s'", field)
		val, err := dest.FieldValueFromData(ctx, field, src)
		if err != nil {
			return err
		}
		logger.Debugf("[portal.chell] sync field '%s' got value '%v'", field, val)
		err = c.dumpField(ctx, val, field)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Chell) dumpAsyncFields(ctx context.Context, src interface{}, dest *Schema) error {
	logger.Debugln("[portal.chell] dump async fields")
	ctx, cancel := context.WithCancel(ctx)
	type Item struct {
		field *Field
		data  interface{}
		err   error
	}

	var wg sync.WaitGroup
	items := make(chan *Item, MinInt(len(dest.AsyncFields()), ConcurrentDumpingPoolSize))

	for _, field := range dest.AsyncFields() {
		wg.Add(1)
		go func(f *Field) {
			defer wg.Done()
			logger.Debugf("[portal.chell] processing sync field '%s'", field)
			val, err := dest.FieldValueFromData(ctx, f, src)
			logger.Debugf("[portal.chell] sync field '%s' got value '%v'", field, val)
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

		err := c.dumpField(ctx, item.data, item.field)
		if err != nil {
			cancel()
			return err
		}
	}

	return nil
}

func (c *Chell) dumpField(ctx context.Context, src interface{}, field *Field) error {
	defer func() {
		if err := recover(); err != nil {
			err = fmt.Sprintf("failed to dump field %s, src is %v: %s", field, src, err)
		}
	}()

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
			return c.dumpFieldNestedMany(ctx, src, field)
		} else {
			logger.Debugf("[portal.chell] dump nested field %s with value %v", field, src)
			return c.dumpFieldNestedOne(ctx, src, field)
		}
	}

	return nil
}

func (c *Chell) dumpFieldNestedOne(ctx context.Context, src interface{}, field *Field) error {
	val := reflect.New(IndirectStructTypeP(reflect.TypeOf(field.Value())))
	toNestedSchema := NewSchema(val.Interface())
	toNestedSchema.SetOnlyFields(field.NestedOnlyNames()...)
	toNestedSchema.SetExcludeFields(field.NestedExcludeNames()...)

	err := c.dump(ctx, src, toNestedSchema)
	if err != nil {
		return err
	}
	switch field.Kind() {
	case reflect.Ptr:
		field.SetValue(val.Interface())
	case reflect.Struct:
		field.SetValue(val.Elem().Interface())
	default:
		panic("invalid nested schema")
	}

	return nil
}

func (c *Chell) dumpFieldNestedMany(ctx context.Context, src interface{}, field *Field) error {
	typ := reflect.TypeOf(field.Value())
	nestedSchemaSlice := reflect.New(typ)

	cpy := c.Only(field.NestedOnlyNames()...).Exclude(field.NestedExcludeNames()...)
	err := cpy.DumpMany(ctx, src, nestedSchemaSlice.Interface())
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

func (c *Chell) MustDump(ctx context.Context, src interface{}, dump interface{}) {
	err := c.Dump(ctx, src, dump)
	if err != nil {
		panic(err)
	}
}

func (c *Chell) DumpMany(ctx context.Context, src interface{}, dest interface{}) error {
	return c.dumpMany(ctx, src, dest)
}

func (c *Chell) dumpMany(ctx context.Context, src interface{}, dest interface{}) error {
	reflectedData := reflect.ValueOf(src)
	if reflectedData.Kind() != reflect.Slice {
		panic("input src must be a slice")
	}

	schemaSlice := reflect.Indirect(reflect.ValueOf(dest))
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
			err := c.dump(ctx, val, toSchema)
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

func (c *Chell) MustDumpMany(ctx context.Context, src interface{}, dest interface{}) {
	err := c.DumpMany(ctx, src, dest)
	if err != nil {
		panic(err)
	}
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
