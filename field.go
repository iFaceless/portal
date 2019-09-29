package portal

import (
	"database/sql/driver"
	"strings"

	"reflect"

	"github.com/fatih/structs"
)

var (
	defaultTagName = "portal"
)

type Field struct {
	*structs.Field
	settings  map[string]string
	isIgnored bool
	schema    *Schema
}

func NewField(schema *Schema, field *structs.Field) *Field {
	schField := &Field{
		Field:    field,
		schema:   schema,
		settings: parseTagSettings(field.Tag(defaultTagName)),
	}

	return schField
}

func (f *Field) String() string {
	return f.schema.Name() + "." + f.Name()
}

func (f *Field) SetValue(v interface{}) error {
	realValue, err := f.realInputValue(v)
	if err != nil {
		return err
	}

	convertedValue, err := Convert(realValue, f.Value())
	if err != nil {
		return f.setIndirectly(realValue)
	} else {
		return f.Set(convertedValue)
	}
}

func (f *Field) realInputValue(v interface{}) (interface{}, error) {
	switch r := v.(type) {
	case driver.Valuer:
		return r.Value()
	default:
		return v, nil
	}
}

func (f *Field) setIndirectly(v interface{}) error {
	outValueType := reflect.TypeOf(f.Value())

	var outValuePtr reflect.Value
	var isFieldValuePtr bool
	if outValueType.Kind() == reflect.Ptr {
		isFieldValuePtr = true
		outValuePtr = reflect.New(outValueType.Elem())
	} else {
		isFieldValuePtr = false
		outValuePtr = reflect.New(outValueType)
	}

	if setter, ok := outValuePtr.Interface().(ValueSetter); ok {
		err := setter.SetValue(v)
		if err != nil {
			return err
		}

		if isFieldValuePtr {
			return f.Field.Set(outValuePtr.Interface())
		} else {
			return f.Field.Set(outValuePtr.Elem().Interface())
		}
	}

	return nil
}

func (f *Field) IsRequired() bool {
	return f.tagHasOption("REQUIRED")
}

func (f *Field) IsNested() bool {
	return f.tagHasOption("NESTED")
}

func (f *Field) Many() bool {
	return f.Kind() == reflect.Slice
}

func (f *Field) Method() string {
	return f.settings["METH"]
}

func (f *Field) HasMethod() bool {
	return f.tagHasOption("METH")
}

func (f *Field) ChainingAttrs() (attrs []string) {
	result, ok := f.settings["ATTR"]
	if !ok {
		return nil
	}
	for _, r := range strings.Split(result, ".") {
		attrs = append(attrs, strings.TrimSpace(r))
	}
	return
}

func (f *Field) HasChainingAttrs() bool {
	return f.tagHasOption("ATTR")
}

func (f *Field) Const() interface{} {
	val, ok := f.settings["CONST"]
	if ok {
		return val
	} else {
		return nil
	}
}

func (f *Field) HasConst() bool {
	return f.tagHasOption("CONST")
}

func (f *Field) tagHasOption(opt string) bool {
	if _, ok := f.settings[opt]; ok {
		return true
	}
	return false
}

func (f *Field) NestedOnlyNames() (names []string) {
	if onlyNames, ok := f.settings["ONLY"]; ok {
		for _, name := range strings.Split(onlyNames, ",") {
			names = append(names, strings.TrimSpace(name))
		}
	}
	return
}

func (f *Field) NestedExcludeNames() (names []string) {
	if excludeNames, ok := f.settings["EXCLUDE"]; ok {
		for _, name := range strings.Split(excludeNames, ",") {
			names = append(names, strings.TrimSpace(name))
		}
	}
	return
}

func (f *Field) Async() bool {
	return f.tagHasOption("ASYNC")
}

func parseTagSettings(s string) map[string]string {
	settings := make(map[string]string)
	for _, item := range strings.Split(s, ";") {
		parts := strings.Split(item, ":")
		if len(parts) > 1 {
			settings[strings.ToUpper(strings.TrimSpace(parts[0]))] = strings.TrimSpace(strings.Join(parts[1:], ":"))
		} else if len(parts) == 1 {
			settings[strings.ToUpper(strings.TrimSpace(parts[0]))] = ""
		}
	}
	return settings
}
