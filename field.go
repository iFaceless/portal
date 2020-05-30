package portal

import (
	"database/sql/driver"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"

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
	return &schemaField{
		Field:    field,
		schema:   schema,
		settings: parseTagSettings(tagStr),
		alias:    parseAlias(field.Tag(schema.fieldAliasMapTagName)),
	}
}

func (f *schemaField) String() string {
	return f.schema.name() + "." + f.Name()
}

// cases:
// value -> value
// value -> *value
// *value -> *value
// *value -> value
// Valuer -> value
// Valuer -> *value
// Valuer -> SetValuer
// value -> SetValuer
func (f *schemaField) setValue(v interface{}) error {
	convertedValue, err := convert(f.Value(), v)
	if err == nil {
		return f.Set(convertedValue)
	}

	indirectValue, err := f.inputValueIndirectly(v)
	if err != nil {
		return errors.WithStack(err)
	}

	convertedValue, err = convert(f.Value(), indirectValue)
	if err == nil {
		return f.Set(convertedValue)
	}

	return f.setIndirectly(indirectValue)
}

func (f *schemaField) inputValueIndirectly(v interface{}) (interface{}, error) {
	var iv interface{}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr:
		iv = v
	default:
		// input is `SomeType`
		// but `*SomeType` implements `Valuer` interface.
		tmpValue := reflect.New(rv.Type())
		tmpValue.Elem().Set(rv)
		iv = tmpValue.Interface()
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

func (f *schemaField) method() (meth string, attrs []string) {
	result, ok := f.settings["METH"]
	if !ok {
		return "", nil
	}

	for _, r := range strings.Split(result, ".") {
		attrs = append(attrs, strings.TrimSpace(r))
	}

	if len(attrs) > 0 {
		meth = attrs[0]
		attrs = attrs[1:]
	}
	return
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

func (f *schemaField) isCacheDisabled() bool {
	return f.tagHasOption("DISABLECACHE")
}

func (f *schemaField) defaultValue() interface{} {
	val, ok := f.settings["DEFAULT"]
	if !ok {
		return nil
	}

	var defaultValue interface{}
	if val == "AUTO_INIT" {
		// just initialize this field, now support ptr/slice/map
		typ := reflect.TypeOf(f.Value())
		switch typ.Kind() {
		case reflect.Ptr:
			defaultValue = reflect.New(typ.Elem()).Interface()
		case reflect.Slice:
			defaultValue = reflect.MakeSlice(typ, 0, 0).Interface()
		case reflect.Map:
			defaultValue = reflect.MakeMap(typ).Interface()
		default:
			defaultValue = reflect.New(typ).Elem().Interface()
		}
		return defaultValue
	}

	return val
}

func (f *schemaField) hasDefaultValue() bool {
	return f.tagHasOption("DEFAULT")
}

func (f *schemaField) tagHasOption(opt string) bool {
	if _, ok := f.settings[opt]; ok {
		return true
	}
	return false
}

func (f *schemaField) nestedOnlyNames(customFilters []*filterNode) (names []string) {
	filterNames := extractFilterNodeNames(
		customFilters, &extractOption{queryByParentName: f.Name(), queryByParentNameAlias: f.alias})
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
		&extractOption{ignoreNodeWithChildren: true, queryByParentName: f.Name(), queryByParentNameAlias: f.alias},
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
	cachedSettings, ok := cachedFieldTagSettings.Load(s)
	if ok {
		result, _ := cachedSettings.(map[string]string)
		return result
	}

	settings := make(map[string]string)
	for _, item := range strings.Split(s, ";") {
		parts := strings.Split(item, ":")
		if len(parts) > 1 {
			settings[strings.ToUpper(strings.TrimSpace(parts[0]))] = strings.TrimSpace(strings.Join(parts[1:], ":"))
		} else if len(parts) == 1 {
			settings[strings.ToUpper(strings.TrimSpace(parts[0]))] = ""
		}
	}

	cachedFieldTagSettings.Store(s, settings)
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
