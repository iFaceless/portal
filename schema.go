package portal

import (
	"fmt"

	"context"

	"github.com/fatih/structs"
)

type Schema struct {
	RawValue            interface{}
	schemaStruct        *structs.Struct
	availableFieldNames map[string]bool
}

func NewSchema(v interface{}) *Schema {
	sch := &Schema{
		schemaStruct:        structs.New(v),
		availableFieldNames: make(map[string]bool, 0),
		RawValue:            v,
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

func (s *Schema) AvailableFields() []*Field {
	fields := make([]*Field, 0)
	for k, v := range s.availableFieldNames {
		if v {
			fields = append(fields, NewField(s, s.schemaStruct.Field(k)))
		}
	}

	return fields
}

func (s *Schema) SyncFields() (fields []*Field) {
	for _, f := range s.AvailableFields() {
		if !f.Async() {
			fields = append(fields, f)
		}
	}
	return
}

func (s *Schema) AsyncFields() (fields []*Field) {
	for _, f := range s.AvailableFields() {
		if f.Async() {
			fields = append(fields, f)
		}
	}
	return
}

func (s *Schema) Struct() *structs.Struct {
	return s.schemaStruct
}

func (s *Schema) FieldValueFromData(ctx context.Context, field *Field, v interface{}) (val interface{}, err error) {
	if IsNil(v) || !structs.IsStruct(v) {
		return nil, fmt.Errorf("failed to get value for field %s, empty input data %v", field, v)
	}

	src := structs.New(v)
	if field.HasConst() {
		val = field.Const()
	} else if field.HasMethod() {
		ret, err := InvokeMethod(ctx, s.RawValue, field.Method(), v)
		if err != nil {
			return nil, fmt.Errorf("failed to get value: %s", err)
		}
		return ret, nil
	} else if field.HasChainingAttrs() {
		val, err = GetNestedValue(ctx, v, field.ChainingAttrs())
	} else {
		if field.IsNested() {
			val, err = GetNestedValue(ctx, v, []string{field.Name()})
		} else {
			f, ok := src.FieldOk(field.Name())
			if ok {
				val = f.Value()
			} else {
				val, err = GetNestedValue(ctx, v, []string{field.Name()})
			}
		}
	}

	return val, nil
}

func (s *Schema) SetOnlyFields(fields ...string) {
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
			panic(fmt.Sprintf("field name '%s.%s' not found", s.Name(), f))
		}
	}
}

func (s *Schema) SetExcludeFields(fields ...string) {
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

// Name 返回 Table 结构体名称
func (s *Schema) Name() string {
	return StructName(s.RawValue)
}
