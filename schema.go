package portal

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fatih/structs"
)

type schema struct {
	fieldAliasMapTagName string
	rawValue             interface{}
	schemaStruct         *structs.Struct
	availableFieldNames  map[string]bool
	fields               []*schemaField
}

func newSchema(v interface{}) *schema {
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
		typ, err := innerStructType(rv.Type())
		if err != nil {
			panic(fmt.Errorf("cannot get schema struct: %s", err))
		}
		schemaValue = reflect.New(typ).Elem()
		rv.Elem().Set(schemaValue.Addr())
	default:
		panic("expect a pointer to struct")
	}

	sch := &schema{
		schemaStruct:         structs.New(schemaValue.Addr().Interface()),
		availableFieldNames:  make(map[string]bool),
		rawValue:             schemaValue.Addr().Interface(),
		fieldAliasMapTagName: "json",
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

func (s *schema) fieldValueFromSrc(ctx context.Context, field *schemaField, v interface{}) (val interface{}, err error) {
	if isNil(v) || !structs.IsStruct(v) {
		return nil, fmt.Errorf("failed to get value for field %s, empty input data %v", field, v)
	}

	src := structs.New(v)
	if field.hasConstValue() {
		val = field.constValue()
	} else if field.hasMethod() {
		ret, err := invokeMethodOfAnyType(ctx, s.rawValue, field.method(), v)
		if err != nil {
			return nil, fmt.Errorf("failed to get value: %s", err)
		}
		return ret, nil
	} else if field.hasChainingAttrs() {
		return nestedValue(ctx, v, field.chainingAttrs())
	} else {
		if field.isNested() {
			return nestedValue(ctx, v, []string{field.Name()})
		} else {
			f, ok := src.FieldOk(field.Name())
			if ok {
				val = f.Value()
			} else {
				v, e := nestedValue(ctx, v, []string{field.Name()})
				val = v
				err = e
			}
		}
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

func (s *schema) fieldByNameOrAlias(name string) *schemaField {
	for _, f := range s.fields {
		if f.alias == name || f.Name() == name {
			return f
		}
	}
	return nil
}
