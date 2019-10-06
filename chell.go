package portal

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/Jeffail/tunny"
	"github.com/pkg/errors"
)

// Chell manages the dumping state and a worker pool.
type Chell struct {
	err error

	onlyFieldFilters    map[int][]*FilterNode
	excludeFieldFilters map[int][]*FilterNode

	// workerPool is a fixed-size goroutine pool to
	// fill data into schema fields concurrently.
	//
	// Since the number of incoming requests are unknown,
	// we must limit the spawned goroutines to avoid
	// consuming too many resources.
	workerPool *tunny.Pool

	// workerPoolSize is default to `runtime.NumCPU()`.
	workerPoolSize int

	// workerTimeout tells the worker how long to wait
	// before getting the processed result.
	// If the worker timeout occurs, Chell will get an
	// error: `ErrJobTimedOut`.
	// Default timeout is '0', which means the worker will
	// wait until the processed result coming out.
	workerTimeout time.Duration
}

// New creates a new Chell instance with a worker pool waiting to be feed.
// After dumping process finished, you're responsible to call the `Close()`
// method to release resource occupied.
// It's highly recommended to call function `portal.Dump()` or
// `portal.DumpWithContext()` directly.
func New(opts ...Option) *Chell {
	chell := &Chell{}
	for _, opt := range opts {
		opt(chell)
	}

	if chell.workerPoolSize <= 0 {
		chell.workerPoolSize = runtime.NumCPU()
	}

	// TODO: custom worker may be required.
	chell.workerPool = tunny.NewFunc(chell.workerPoolSize, func(i interface{}) interface{} {
		return ""
	})

	return chell
}

func Dump(dst, src interface{}, opts ...Option) error {
	return DumpWithContext(context.TODO(), dst, src, opts...)
}

func DumpWithContext(ctx context.Context, dst, src interface{}, opts ...Option) error {
	chell := New(opts...)
	defer chell.Close()
	return chell.DumpWithContext(ctx, dst, src)
}

func (c *Chell) Dump(dst, src interface{}) error {
	return c.DumpWithContext(context.TODO(), dst, src)
}

func (c *Chell) DumpWithContext(ctx context.Context, dst, src interface{}) error {
	if c.err != nil {
		return errors.WithStack(c.err)
	}

	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr {
		return errors.New("dst must be a pointer")
	}

	if reflect.Indirect(rv).Kind() == reflect.Slice {
		return c.dumpMany(
			ctx, dst, src,
			ExtractFilterNodeNames(c.onlyFieldFilters[0], nil),
			ExtractFilterNodeNames(c.excludeFieldFilters[0], &ExtractOption{ignoreNodeWithChildren: true}))
	} else {
		toSchema := NewSchema(dst)
		toSchema.SetOnlyFields(ExtractFilterNodeNames(c.onlyFieldFilters[0], nil)...)
		toSchema.SetExcludeFields(ExtractFilterNodeNames(c.excludeFieldFilters[0], &ExtractOption{ignoreNodeWithChildren: true})...)
		return c.dump(IncrDumpDepthContext(ctx), toSchema, src)
	}
}

// Close closes the inner dumping goroutine pool and terminate all the dumping workers.
func (c *Chell) Close() {
	if c.workerPool != nil {
		c.workerPool.Close()
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
	items := make(chan *Item, MinInt(len(dst.AsyncFields()), 10))

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
			return item.err
		}

		err := c.dumpField(ctx, item.field, item.data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Chell) dumpField(ctx context.Context, field *Field, value interface{}) error {
	if IsNil(value) {
		logger.Warnf("[portal.chell] cannot get value for field %s, current input value is %v", field, value)
		return nil
	}

	if Convertible(value, field.Field.Value()) {
		return field.SetValue(value)
	}
	if !field.IsNested() {
		logger.Debugf("[portal.chell] dump normal field %s with value %v", field, value)
		return field.SetValue(value)
	} else {
		if field.Many() {
			logger.Debugf("[portal.chell] dump nested slice field %s with value %v", field, value)
			return c.dumpFieldNestedMany(ctx, field, value)
		} else {
			logger.Debugf("[portal.chell] dump nested field %s with value %v", field, value)
			return c.dumpFieldNestedOne(ctx, field, value)
		}
	}
}

func (c *Chell) dumpFieldNestedOne(ctx context.Context, field *Field, src interface{}) error {
	val := reflect.New(IndirectStructTypeP(reflect.TypeOf(field.Value())))
	toNestedSchema := NewSchema(val.Interface())

	depth := DumpDepthFromContext(ctx)
	toNestedSchema.SetOnlyFields(field.NestedOnlyNames(c.onlyFieldFilters[depth])...)
	toNestedSchema.SetExcludeFields(field.NestedExcludeNames(c.excludeFieldFilters[depth])...)
	err := c.dump(IncrDumpDepthContext(ctx), toNestedSchema, src)
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
	depth := DumpDepthFromContext(ctx)
	err := c.dumpMany(
		ctx,
		nestedSchemaSlice.Interface(),
		src,
		field.NestedOnlyNames(c.onlyFieldFilters[depth]),
		field.NestedExcludeNames(c.excludeFieldFilters[depth]),
	)
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

func (c *Chell) dumpMany(ctx context.Context, dst, src interface{}, onlyFields, excludeFields []string) error {
	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Ptr {
		rv = reflect.Indirect(rv)
	}

	if rv.Kind() != reflect.Slice {
		panic("input src must be a slice")
	}

	schemaSlice := reflect.Indirect(reflect.ValueOf(dst))
	schemaSlice.Set(reflect.MakeSlice(schemaSlice.Type(), rv.Len(), rv.Cap()))

	schemaType := IndirectStructTypeP(schemaSlice.Type())

	var wg sync.WaitGroup

	type Item struct {
		index     int
		schemaPtr reflect.Value
		err       error
	}

	items := make(chan *Item, MinInt(rv.Len(), 10))

	for i := 0; i < rv.Len(); i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			schemaPtr := reflect.New(schemaType)
			toSchema := NewSchema(schemaPtr.Interface())
			toSchema.SetOnlyFields(onlyFields...)
			toSchema.SetExcludeFields(excludeFields...)
			val := rv.Index(index).Interface()
			err := c.dump(IncrDumpDepthContext(ctx), toSchema, val)
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
	filters, err := ParseFilters(fields)
	if err != nil {
		c.err = err
	} else {
		c.onlyFieldFilters = filters
	}
	return c
}

func (c *Chell) Exclude(fields ...string) *Chell {
	filters, err := ParseFilters(fields)
	if err != nil {
		c.err = err
	} else {
		c.excludeFieldFilters = filters
	}
	return c
}
