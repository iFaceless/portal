package portal

import (
	"context"
	"fmt"
	"reflect"

	"github.com/fatih/structs"
)

type schema struct {
	RawValue            interface{}
	schemaStruct        *structs.Struct
	availableFieldNames map[string]bool
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
		schemaStruct:        structs.New(schemaValue.Addr().Interface()),
		availableFieldNames: make(map[string]bool),
		RawValue:            schemaValue.Addr().Interface(),
	}

	for _, name := range getAvailableFieldNames(sch.schemaStruct.Fields()) {
		sch.availableFieldNames[name] = true
	}

	return sch
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
	for k, v := range s.availableFieldNames {
		if v {
			fields = append(fields, newField(s, s.schemaStruct.Field(k)))
		}
	}

	return fields
}

func (s *schema) syncFields() (fields []*schemaField) {
	for _, f := range s.availableFields() {
		if !f.async() {
			fields = append(fields, f)
		}
	}
	return
}

func (s *schema) asyncFields() (fields []*schemaField) {
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
		ret, err := invokeStructMethod(ctx, s.RawValue, field.method(), v)
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

func (s *schema) setOnlyFields(fields ...string) {
	if len(fields) == 0 {
		return
	}

	for k := range s.availableFieldNames {
		s.availableFieldNames[k] = false
	}

	for _, f := range fields {
		if _, ok := s.availableFieldNames[f]; ok {
			s.availableFieldNames[f] = true
		} else {
			panic(fmt.Sprintf("field name '%s.%s' not found", s.name(), f))
		}
	}
}

func (s *schema) setExcludeFields(fields ...string) {
	if len(fields) == 0 {
		return
	}

	for _, f := range fields {
		if _, ok := s.availableFieldNames[f]; ok {
			s.availableFieldNames[f] = false
		} else {
			panic(fmt.Sprintf("field name '%s' not found", f))
		}
	}
}

func (s *schema) name() string {
	return structName(s.RawValue)
}
