package portal

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// Chell manages the dumping state.
type Chell struct {
	// json, yaml etc.
	fieldAliasMapTagName string
	disableConcurrency   bool
	disableCache         bool
	onlyFieldFilters     map[int][]*filterNode
	excludeFieldFilters  map[int][]*filterNode

	// custom field tags
	customFieldTagMap map[string]string
}

// New creates a new Chell instance with a worker pool waiting to be feed.
// It's highly recommended to call function `portal.Dump()` or
// `portal.DumpWithContext()` directly.
func New(opts ...option) (*Chell, error) {
	chell := &Chell{
		fieldAliasMapTagName: "json",
	}

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
func Dump(dst, src interface{}, opts ...option) error {
	return DumpWithContext(context.TODO(), dst, src, opts...)
}

// DumpWithContext dumps src data to dst with an extra context param.
// You can filter fields with optional config `portal.Only` or `portal.Exclude`.
func DumpWithContext(ctx context.Context, dst, src interface{}, opts ...option) error {
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
			extractFilterNodeNames(c.onlyFieldFilters[0], nil),
			extractFilterNodeNames(c.excludeFieldFilters[0], &extractOption{ignoreNodeWithChildren: true}),
			"",
		)
	} else {
		toSchema := newSchema(dst).withFieldAliasMapTagName(c.fieldAliasMapTagName)
		toSchema.setOnlyFields(extractFilterNodeNames(c.onlyFieldFilters[0], nil)...)
		toSchema.setExcludeFields(extractFilterNodeNames(c.excludeFieldFilters[0], &extractOption{ignoreNodeWithChildren: true})...)
		return c.dump(incrDumpDepthContext(ctx), toSchema, src)
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
	filters, err := parseFilters(fields)
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
	filters, err := parseFilters(fields)
	if err != nil {
		return errors.WithStack(err)
	} else {
		c.excludeFieldFilters = filters
	}
	return nil
}

func (c *Chell) dump(ctx context.Context, dst *schema, src interface{}) error {
	// read custom field tags
	for _, field := range dst.fields {
		key := fmt.Sprintf("%s.%s", field.schema.name(), field.Name())
		if v, ok := c.customFieldTagMap[key]; ok {
			field.settings = parseTagSettings(v)
		}
	}

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

func (c *Chell) dumpSyncFields(ctx context.Context, dst *schema, src interface{}) error {
	syncFields := dst.syncFields(c.disableConcurrency)
	if len(syncFields) == 0 {
		return nil
	}

	logger.Debugf("[portal.chell] dump sync fields: %s", syncFields)
	for _, field := range syncFields {
		logger.Debugf("[portal.chell] processing sync field '%s'", field)
		val, err := dst.fieldValueFromSrc(ctx, field, src, c.disableCache)
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

func (c *Chell) dumpAsyncFields(ctx context.Context, dst *schema, src interface{}) error {
	asyncFields := dst.asyncFields(c.disableConcurrency)
	if len(asyncFields) == 0 {
		return nil
	}

	logger.Debugf("[portal.chell] dump async fields: %s", asyncFields)
	type Result struct {
		field *schemaField
		data  interface{}
	}

	type Payload struct {
		field *schemaField
	}

	workerPayloads := make([]interface{}, 0, len(asyncFields))
	for _, field := range asyncFields {
		workerPayloads = append(workerPayloads, &Payload{field: field})
	}

	jobResults, err := submitJobs(
		ctx,
		func(payload interface{}) (interface{}, error) {
			p := payload.(*Payload)
			logger.Debugf("[portal.chell] processing async field '%s'", p.field)
			val, err := dst.fieldValueFromSrc(ctx, p.field, src, c.disableCache)
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

func (c *Chell) dumpField(ctx context.Context, field *schemaField, value interface{}) error {
	if isNil(value) {
		if field.hasDefaultValue() {
			value = field.defaultValue()
			logger.Infof("[portal.chell] use default value for field `%s`", field)
		} else {
			logger.Warnf("[portal.chell] cannot get value for field `%s`, current input value is %v", field, value)
			return nil
		}
	}

	if !field.isNested() {
		logger.Debugf("[portal.chell] dump normal field %s with value '%v'", field, value)
		return field.setValue(value)
	} else {
		if field.hasMany() {
			logger.Debugf("[portal.chell] dump nested slice field %s with value '%v'", field, value)
			return c.dumpFieldNestedMany(ctx, field, value)
		} else {
			logger.Debugf("[portal.chell] dump nested field %s with value '%v'", field, value)
			return c.dumpFieldNestedOne(ctx, field, value)
		}
	}
}

func (c *Chell) dumpFieldNestedOne(ctx context.Context, field *schemaField, src interface{}) error {
	val := reflect.New(indirectStructTypeP(reflect.TypeOf(field.Value())))
	toNestedSchema := newSchema(val.Interface()).withFieldAliasMapTagName(c.fieldAliasMapTagName)

	depth := dumpDepthFromContext(ctx)
	toNestedSchema.setOnlyFields(field.nestedOnlyNames(c.onlyFieldFilters[depth])...)
	toNestedSchema.setExcludeFields(field.nestedExcludeNames(c.excludeFieldFilters[depth])...)
	err := c.dump(incrDumpDepthContext(ctx), toNestedSchema, src)
	if err != nil {
		return err
	}
	switch field.Kind() {
	case reflect.Ptr:
		return field.setValue(val.Interface())
	case reflect.Struct:
		return field.setValue(val.Elem().Interface())
	default:
		panic("invalid nested schema")
	}
}

func (c *Chell) dumpFieldNestedMany(ctx context.Context, field *schemaField, src interface{}) error {
	typ := reflect.TypeOf(field.Value())
	nestedSchemaSlice := reflect.New(typ)
	depth := dumpDepthFromContext(ctx)
	err := c.dumpMany(
		ctx,
		nestedSchemaSlice.Interface(),
		src,
		field.nestedOnlyNames(c.onlyFieldFilters[depth]),
		field.nestedExcludeNames(c.excludeFieldFilters[depth]),
		field.String(),
	)
	if err != nil {
		return err
	}

	switch typ.Kind() {
	case reflect.Ptr:
		err = field.setValue(nestedSchemaSlice.Interface())
	case reflect.Slice:
		err = field.setValue(nestedSchemaSlice.Elem().Interface())
	default:
		panic("invalid nested schema")
	}
	if err != nil {
		return err
	}

	return nil
}

func (c *Chell) dumpMany(ctx context.Context, dst, src interface{}, onlyFields, excludeFields []string, field string) error {
	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Ptr {
		rv = reflect.Indirect(rv)
	}

	if rv.Kind() != reflect.Slice {
		if field != "" {
			panic(fmt.Sprintf("input src must be a slice, current processing field is `%s`", field))
		} else {
			panic("input src must be a slice")
		}
	}

	schemaSlice := reflect.Indirect(reflect.ValueOf(dst))
	schemaSlice.Set(reflect.MakeSlice(schemaSlice.Type(), rv.Len(), rv.Cap()))
	schemaType := indirectStructTypeP(schemaSlice.Type())

	if c.disableConcurrency || !hasAsyncFields(schemaType, onlyFields, excludeFields) {
		return c.dumpManySynchronously(ctx, schemaType, schemaSlice, rv, onlyFields, excludeFields)
	}

	return c.dumpManyConcurrently(ctx, schemaType, schemaSlice, rv, onlyFields, excludeFields)
}

func (c *Chell) dumpManySynchronously(ctx context.Context, schemaType reflect.Type, dst, src reflect.Value, onlyFields, excludeFields []string) error {
	logger.Debugf("[portal.dumpManySynchronously] '%s' -> '%s'", src.Type().String(), dst.Type().String())
	for i := 0; i < src.Len(); i++ {
		schemaPtr := reflect.New(schemaType)
		toSchema := newSchema(schemaPtr.Interface()).withFieldAliasMapTagName(c.fieldAliasMapTagName)
		toSchema.setOnlyFields(onlyFields...)
		toSchema.setExcludeFields(excludeFields...)
		val := src.Index(i).Interface()
		err := c.dump(incrDumpDepthContext(ctx), toSchema, val)
		if err != nil {
			return errors.WithStack(err)
		}

		elem := dst.Index(i)
		switch elem.Kind() {
		case reflect.Struct:
			elem.Set(reflect.Indirect(schemaPtr))
		case reflect.Ptr:
			elem.Set(schemaPtr)
		default:
			return errors.Errorf("unsupported schema field type '%s', expected a struct or a pointer to struct", elem.Type().Kind())
		}
	}
	return nil
}

func (c *Chell) dumpManyConcurrently(ctx context.Context, schemaType reflect.Type, dst, src reflect.Value, onlyFields, excludeFields []string) error {
	logger.Debugf("[portal.dumpManyConcurrently] '%s' -> '%s'", src.Type().String(), dst.Type().String())
	type Result struct {
		index     int
		schemaPtr reflect.Value
	}

	payloads := make([]interface{}, 0, src.Len())
	for i := 0; i < src.Len(); i++ {
		payloads = append(payloads, i)
	}

	jobResults, err := submitJobs(
		ctx,
		func(payload interface{}) (interface{}, error) {
			index := payload.(int)
			schemaPtr := reflect.New(schemaType)
			toSchema := newSchema(schemaPtr.Interface()).withFieldAliasMapTagName(c.fieldAliasMapTagName)
			toSchema.setOnlyFields(onlyFields...)
			toSchema.setExcludeFields(excludeFields...)
			val := src.Index(index).Interface()
			err := c.dump(incrDumpDepthContext(ctx), toSchema, val)
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
		elem := dst.Index(r.index)
		switch elem.Kind() {
		case reflect.Struct:
			elem.Set(reflect.Indirect(r.schemaPtr))
		case reflect.Ptr:
			elem.Set(r.schemaPtr)
		}
	}
	return nil
}
