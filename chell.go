package portal

import (
	"context"
	"reflect"

	"github.com/pkg/errors"
)

// Chell manages the dumping state.
type Chell struct {
	onlyFieldFilters    map[int][]*FilterNode
	excludeFieldFilters map[int][]*FilterNode
}

// New creates a new Chell instance with a worker pool waiting to be feed.
// It's highly recommended to call function `portal.Dump()` or
// `portal.DumpWithContext()` directly.
func New(opts ...Option) (*Chell, error) {
	chell := &Chell{}
	for _, opt := range opts {
		err := opt(chell)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return chell, nil
}

// Dump dumps src data to dst. You can filter fields with
// optional config `portal.Only` or `portal.Exclude`.
func Dump(dst, src interface{}, opts ...Option) error {
	return DumpWithContext(context.TODO(), dst, src, opts...)
}

// DumpWithContext dumps src data to dst with an extra context param.
// You can filter fields with optional config `portal.Only` or `portal.Exclude`.
func DumpWithContext(ctx context.Context, dst, src interface{}, opts ...Option) error {
	chell, err := New(opts...)
	if err != nil {
		return errors.WithStack(err)
	}

	return chell.DumpWithContext(ctx, dst, src)
}

// Dump dumps src data to dst. You can filter fields with
// optional config `portal.Only` or `portal.Exclude`.
func (c *Chell) Dump(dst, src interface{}) error {
	return c.DumpWithContext(context.TODO(), dst, src)
}

// DumpWithContext dumps src data to dst with an extra context param.
// You can filter fields with optional config `portal.Only` or `portal.Exclude`.
func (c *Chell) DumpWithContext(ctx context.Context, dst, src interface{}) error {
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

// SetOnlyFields specifies the fields to keep.
// Examples:
// ```
// c := New()
// c.SetOnlyFields("A") // keep field A only
// c.SetOnlyFields("A[B,C]") // keep field B and C of the nested struct A
// ```
func (c *Chell) SetOnlyFields(fields ...string) error {
	filters, err := ParseFilters(fields)
	if err != nil {
		return errors.WithStack(err)
	} else {
		c.onlyFieldFilters = filters
	}
	return nil
}

// SetOnlyFields specifies the fields to exclude.
// Examples:
// ```
// c := New()
// c.SetExcludeFields("A") // exclude field A
// c.SetExcludeFields("A[B,C]") // exclude field B and C of the nested struct A, but other fields of struct A are still selected.
// ```
func (c *Chell) SetExcludeFields(fields ...string) error {
	filters, err := ParseFilters(fields)
	if err != nil {
		return errors.WithStack(err)
	} else {
		c.excludeFieldFilters = filters
	}
	return nil
}

func (c *Chell) dump(ctx context.Context, dst *Schema, src interface{}) error {
	err := c.dumpSyncFields(ctx, dst, src)
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.dumpAsyncFields(ctx, dst, src)
	if err != nil {
		return errors.WithStack(err)
	}
	return err
}

func (c *Chell) dumpSyncFields(ctx context.Context, dst *Schema, src interface{}) error {
	syncFields := dst.SyncFields()
	if len(syncFields) == 0 {
		return nil
	}

	logger.Debugf("[portal.chell] dump sync fields: %s", syncFields)
	for _, field := range syncFields {
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
	asyncFields := dst.AsyncFields()
	if len(asyncFields) == 0 {
		return nil
	}

	logger.Debugf("[portal.chell] dump async fields: %s", asyncFields)
	type Result struct {
		field *Field
		data  interface{}
	}

	type Payload struct {
		field *Field
	}

	workerPayloads := make([]interface{}, 0, len(asyncFields))
	for _, field := range asyncFields {
		workerPayloads = append(workerPayloads, &Payload{field: field})
	}

	jobResults, err := SubmitJobs(
		ctx,
		func(payload interface{}) (interface{}, error) {
			p := payload.(*Payload)
			logger.Debugf("[portal.chell] processing async field '%s'", p.field)
			val, err := dst.FieldValueFromData(ctx, p.field, src)
			logger.Debugf("[portal.chell] async field '%s' got value '%v'", p.field, val)
			return &Result{field: p.field, data: val}, err
		},
		workerPayloads...)
	if err != nil {
		return errors.WithStack(err)
	}

	for jobResult := range jobResults {
		if jobResult.Err != nil {
			return errors.WithStack(jobResult.Err)
		}

		result := jobResult.Data.(*Result)
		e := c.dumpField(ctx, result.field, result.data)
		if e != nil {
			return errors.WithStack(err)
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

	type Result struct {
		index     int
		schemaPtr reflect.Value
		err       error
	}

	payloads := make([]interface{}, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		payloads = append(payloads, i)
	}

	jobResults, err := SubmitJobs(
		ctx,
		func(payload interface{}) (interface{}, error) {
			index := payload.(int)
			schemaPtr := reflect.New(schemaType)
			toSchema := NewSchema(schemaPtr.Interface())
			toSchema.SetOnlyFields(onlyFields...)
			toSchema.SetExcludeFields(excludeFields...)
			val := rv.Index(index).Interface()
			err := c.dump(IncrDumpDepthContext(ctx), toSchema, val)
			return &Result{index: index, schemaPtr: schemaPtr}, err
		},
		payloads...)
	if err != nil {
		return errors.WithStack(err)
	}

	for jobResult := range jobResults {
		if jobResult.Err != nil {
			return errors.WithStack(jobResult.Err)
		}

		r := jobResult.Data.(*Result)
		elem := schemaSlice.Index(r.index)
		switch elem.Kind() {
		case reflect.Struct:
			elem.Set(reflect.Indirect(r.schemaPtr))
		case reflect.Ptr:
			elem.Set(r.schemaPtr)
		}
	}
	return nil
}
