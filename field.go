package portal

import (
	"database/sql/driver"
	"regexp"
	"strings"
	"sync"

	"reflect"

	"github.com/fatih/structs"
)

var (
	defaultTagName         = "portal"
	cachedFieldTagSettings sync.Map
	cachedAliasMap         sync.Map
)

type schemaField struct {
	*structs.Field
	settings  map[string]string
	isIgnored bool //nolint
	schema    *schema
	alias     string
}

func newField(schema *schema, field *structs.Field) *schemaField {
	tagStr := field.Tag(defaultTagName)

	var settings map[string]string
	cachedSettings, ok := cachedFieldTagSettings.Load(tagStr)
	if ok {
		result, _ := cachedSettings.(map[string]string)
		settings = result
	} else {
		result := parseTagSettings(tagStr)
		cachedFieldTagSettings.Store(tagStr, result)
		settings = result
	}

	return &schemaField{
		Field:    field,
		schema:   schema,
		settings: settings,
		alias:    parseAlias(field.Tag(schema.fieldAliasMapTagName)),
	}
}

func (f *schemaField) String() string {
	return f.schema.name() + "." + f.Name()
}

func (f *schemaField) setValue(v interface{}) error {
	realValue, err := f.realInputValue(v)
	if err != nil {
		return err
	}

	convertedValue, err := convert(f.Value(), realValue)
	if err != nil {
		return f.setIndirectly(realValue)
	} else {
		return f.Set(convertedValue)
	}
}

func (f *schemaField) realInputValue(v interface{}) (interface{}, error) {
	var iv interface{}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr:
		iv = v
	case reflect.Struct:
		// input is `SomeStruct`
		// but `*SomeStruct` implements `Valuer` interface.
		tmpValue := reflect.New(rv.Type())
		tmpValue.Elem().Set(rv)
		iv = tmpValue.Interface()
	default:
		return v, nil
	}

	switch r := iv.(type) {
	case driver.Valuer:
		return r.Value()
	case Valuer:
		return r.Value()
	default:
		return v, nil
	}
}

func (f *schemaField) setIndirectly(v interface{}) error {
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

func (f *schemaField) isRequired() bool {
	return f.tagHasOption("REQUIRED")
}

func (f *schemaField) isNested() bool {
	return f.tagHasOption("NESTED")
}

func (f *schemaField) hasMany() bool {
	return f.Kind() == reflect.Slice
}

func (f *schemaField) method() string {
	return f.settings["METH"]
}

func (f *schemaField) hasMethod() bool {
	return f.tagHasOption("METH")
}

func (f *schemaField) chainingAttrs() (attrs []string) {
	result, ok := f.settings["ATTR"]
	if !ok {
		return nil
	}
	for _, r := range strings.Split(result, ".") {
		attrs = append(attrs, strings.TrimSpace(r))
	}
	return
}

func (f *schemaField) hasChainingAttrs() bool {
	return f.tagHasOption("ATTR")
}

func (f *schemaField) constValue() interface{} {
	val, ok := f.settings["CONST"]
	if ok {
		return val
	} else {
		return nil
	}
}

func (f *schemaField) hasConstValue() bool {
	return f.tagHasOption("CONST")
}

func (f *schemaField) tagHasOption(opt string) bool {
	if _, ok := f.settings[opt]; ok {
		return true
	}
	return false
}

func (f *schemaField) nestedOnlyNames(customFilters []*filterNode) (names []string) {
	filterNames := extractFilterNodeNames(
		customFilters, &extractOption{queryByParentName: f.Name()})
	if len(filterNames) > 0 {
		return filterNames
	} else {
		return f.nestedOnlyNamesParsedFromTag()
	}
}

func (f *schemaField) nestedOnlyNamesParsedFromTag() (names []string) {
	if onlyNames, ok := f.settings["ONLY"]; ok {
		for _, name := range strings.Split(onlyNames, ",") {
			names = append(names, strings.TrimSpace(name))
		}
	}
	return
}

func (f *schemaField) nestedExcludeNames(customFilters []*filterNode) []string {
	fieldNames := extractFilterNodeNames(
		customFilters,
		&extractOption{ignoreNodeWithChildren: true, queryByParentName: f.Name()},
	)
	if len(fieldNames) > 0 {
		return fieldNames
	} else {
		return f.nestedExcludeNamesParsedFromTag()
	}
}

func (f *schemaField) nestedExcludeNamesParsedFromTag() (names []string) {
	if excludeNames, ok := f.settings["EXCLUDE"]; ok {
		for _, name := range strings.Split(excludeNames, ",") {
			names = append(names, strings.TrimSpace(name))
		}
	}
	return
}

func (f *schemaField) async() bool {
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

func parseAlias(s string) string {
	ret, ok := cachedAliasMap.Load(s)
	if ok {
		return ret.(string)
	}

	parts := strings.Split(s, ",")
	if len(parts) == 0 {
		cachedAliasMap.Store(s, "")
		return ""
	}
	alias := strings.TrimSpace(parts[0])
	re, err := regexp.Compile(`^[_a-zA-Z0-9]+-*[_a-zA-Z0-9]+$`)
	if err != nil {
		cachedAliasMap.Store(s, "")
		return ""
	}
	if re.MatchString(alias) {
		cachedAliasMap.Store(s, alias)
		return alias
	}

	cachedAliasMap.Store(s, "")
	return ""
}
