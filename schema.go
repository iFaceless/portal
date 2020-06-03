package portal

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/structs"
)

type schema struct {
	fieldAliasMapTagName string
	rawValue             interface{}
	schemaStruct         *structs.Struct
	availableFieldNames  map[string]bool
	fields               []*schemaField

	parent *schema

	cacheDisabled bool
	cacheGroup    *cacheGroup
}

func newSchema(v interface{}, parent ...*schema) *schema {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		panic("expect a pointer to struct")
	}

	var schemaValue reflect.Value
	switch rv.Elem().Kind() {
	case reflect.Struct:
		// var schema SchemaStruct
		// ptr := &schema
		schemaValue = rv.Elem()
	case reflect.Ptr:
		// var schema *SchemaStruct
		// ptr := &schema
		if rv.Elem().IsNil() {
			typ, err := innerStructType(rv.Type())
			if err != nil {
				panic(fmt.Errorf("cannot get schema struct: %s", err))
			}
			schemaValue = reflect.New(typ).Elem()
			rv.Elem().Set(schemaValue.Addr())
		} else {
			schemaValue = rv.Elem().Elem()
		}
	default:
		panic("expect a pointer to struct")
	}

	rawValue := schemaValue.Addr().Interface()

	var cacheDisabled = func(in interface{}) bool {
		if ret, ok := in.(cachable); ok {
			return ret.PortalDisableCache()
		}
		return false
	}(rawValue)

	sch := &schema{
		schemaStruct:         structs.New(schemaValue.Addr().Interface()),
		availableFieldNames:  make(map[string]bool),
		rawValue:             rawValue,
		fieldAliasMapTagName: "json",
		cacheDisabled:        cacheDisabled,
		cacheGroup:           newCacheGroup(newMapCache()),
	}

	if len(parent) > 0 {
		sch.parent = parent[0]
	}

	for _, name := range getAvailableFieldNames(sch.schemaStruct.Fields()) {
		sch.availableFieldNames[name] = true
		sch.fields = append(sch.fields, newField(sch, sch.schemaStruct.Field(name)))
	}

	return sch
}

func hasAsyncFields(schemaType reflect.Type, onlyFields, excludeFields []string) bool {
	// TODO: try to cache the result
	schema := newSchema(reflect.New(schemaType).Interface())
	schema.setOnlyFields(onlyFields...)
	schema.setExcludeFields(excludeFields...)
	return len(schema.asyncFields(false)) > 0
}

func (s *schema) withFieldAliasMapTagName(t string) *schema {
	s.fieldAliasMapTagName = t
	return s
}

func getAvailableFieldNames(fields []*structs.Field) (names []string) {
	for _, f := range fields {
		if f.IsEmbedded() {
			names = append(names, getAvailableFieldNames(f.Fields())...)
		} else {
			names = append(names, f.Name())
		}
	}
	return names
}

func (s *schema) availableFields() []*schemaField {
	fields := make([]*schemaField, 0)
	for _, f := range s.fields {
		v, ok := s.availableFieldNames[f.Name()]
		if ok && v {
			fields = append(fields, f)
		}
	}
	return fields
}

func (s *schema) syncFields(disableConcurrency bool) (fields []*schemaField) {
	for _, f := range s.availableFields() {
		if disableConcurrency {
			fields = append(fields, f)
		} else if !f.async() {
			fields = append(fields, f)
		}
	}
	return
}

func (s *schema) asyncFields(disableConcurrency bool) (fields []*schemaField) {
	if disableConcurrency {
		return
	}

	for _, f := range s.availableFields() {
		if f.async() {
			fields = append(fields, f)
		}
	}
	return
}

func (s *schema) innerStruct() *structs.Struct {
	return s.schemaStruct
}

func (s *schema) fieldValueFromSrc(ctx context.Context, field *schemaField, v interface{}, noCache bool) (val interface{}, err error) {
	if isNil(v) || !structs.IsStruct(v) {
		return nil, fmt.Errorf("failed to get value for field %s, empty input data %v", field, v)
	}

	if field.hasConstValue() {
		val = field.constValue()
	} else if field.hasMethod() {
		m, attrs := field.method()
		if m == "" {
			return nil, fmt.Errorf("empty method name")
		}

		var ret interface{}
		var err error
		disableCache := noCache || field.isCacheDisabled()
		if disableCache {
			ret, err = invokeMethodOfAnyType(ctx, s.rawValue, m, v)
		} else {
			cacheKey := genCacheKey(ctx, s.rawValue, v, m)
			ret, err = invokeMethodOfAnyTypeWithCache(ctx, s.rawValue, m, s.cacheGroup, cacheKey, v)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get value: %s", err)
		}
		if len(attrs) > 0 {
			return nestedValue(ctx, ret, attrs, nil, !disableCache)
		}
		return ret, nil
	} else if field.hasChainingAttrs() {
		disableCache := noCache || field.isCacheDisabled()
		return nestedValue(ctx, v, field.chainingAttrs(), s.cacheGroup, !disableCache)
	} else {
		return nestedValue(ctx, v, []string{field.Name()}, nil, false)
	}

	return
}

func (s *schema) setOnlyFields(fieldNames ...string) {
	if len(fieldNames) == 0 {
		return
	}

	for k := range s.availableFieldNames {
		s.availableFieldNames[k] = false
	}

	for _, f := range fieldNames {
		field := s.fieldByNameOrAlias(f)
		if field == nil {
			logger.Warnf("field name '%s.%s' not found", s.name(), f)
		} else {
			s.availableFieldNames[field.Name()] = true
		}
	}
}

func (s *schema) setExcludeFields(fieldNames ...string) {
	if len(fieldNames) == 0 {
		return
	}

	for _, f := range fieldNames {
		field := s.fieldByNameOrAlias(f)
		if field == nil {
			logger.Warnf("field name '%s.%s' not found", s.name(), f)
		} else {
			s.availableFieldNames[field.Name()] = false
		}
	}
}

func (s *schema) name() string {
	return structName(s.rawValue)
}

func (s *schema) nameWithParents() string {
	var names []string

	p := s
	for p != nil {
		names = append(names, p.name())
		p = p.parent
	}

	// reverse names (two cursors)
	i, j := 0, len(names)-1
	for i < j {
		names[i], names[j] = names[j], names[i]
		i++
		j--
	}

	return strings.Join(names, ".")
}

func (s *schema) fieldByNameOrAlias(name string) *schemaField {
	for _, f := range s.fields {
		if f.alias == name || f.Name() == name {
			return f
		}
	}
	return nil
}
