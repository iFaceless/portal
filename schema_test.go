package portal

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SchoolSchema struct {
	Name string
	Addr string
}

type PersonSchema struct {
	ID  string `json:"id"`
	Age int    `json:"age"`
}

type UserSchema2 struct {
	PersonSchema
	Name   string        `portal:"meth:GetName"`
	School *SchoolSchema `portal:"nested"`
	Async  int           `json:"async" portal:"async"`
}

func (u *UserSchema2) GetName(user interface{}) interface{} {
	return "test"
}

func TestSchema(t *testing.T) {
	schema := newSchema(&UserSchema2{})
	assert.Equal(t, "UserSchema2", schema.name())
	assert.NotNil(t, schema.innerStruct())

	user := struct {
		ID     int
		School *SchoolSchema
	}{10, &SchoolSchema{Name: "test school"}}

	idField := newField(schema, schema.schemaStruct.Field("ID"))
	val, err := schema.fieldValueFromSrc(context.TODO(), idField, user)
	assert.Nil(t, err)
	assert.Equal(t, 10, val)

	nameField := newField(schema, schema.schemaStruct.Field("Name"))
	val, err = schema.fieldValueFromSrc(context.TODO(), nameField, user)
	assert.Nil(t, err)
	assert.Equal(t, "test", val)

	schoolField := newField(schema, schema.schemaStruct.Field("School"))
	val, err = schema.fieldValueFromSrc(context.TODO(), schoolField, user)
	assert.Nil(t, err)
	assert.Equal(t, &SchoolSchema{Name: "test school", Addr: ""}, val)
}

func Test_hasAsyncFields(t *testing.T) {
	assert.False(t, hasAsyncFields(reflect.TypeOf(&PersonSchema{}), nil, nil))
	assert.False(t, hasAsyncFields(reflect.TypeOf(PersonSchema{}), nil, nil))
	assert.True(t, hasAsyncFields(reflect.TypeOf(UserSchema2{}), nil, nil))
	assert.True(t, hasAsyncFields(reflect.TypeOf(&UserSchema2{}), nil, nil))
	assert.True(t, hasAsyncFields(reflect.TypeOf(&UserSchema2{}), []string{"async"}, nil))
	assert.True(t, hasAsyncFields(reflect.TypeOf(&UserSchema2{}), []string{"Async"}, nil))
	assert.False(t, hasAsyncFields(reflect.TypeOf(&UserSchema2{}), []string{"Name"}, nil))
	assert.False(t, hasAsyncFields(reflect.TypeOf(&UserSchema2{}), []string{}, []string{"Async"}))
	assert.False(t, hasAsyncFields(reflect.TypeOf(&UserSchema2{}), []string{}, []string{"async"}))
}

func TestSchema_GetFields(t *testing.T) {
	schema := newSchema(&UserSchema2{}).withFieldAliasMapTagName("json")
	assert.ElementsMatch(t, []string{"Age", "ID", "Name", "School", "Async"}, filedNames(schema.availableFields()))

	assert.ElementsMatch(t, []string{"Age", "ID", "Name", "School"}, filedNames(schema.syncFields(false)))
	assert.ElementsMatch(t, []string{"Async"}, filedNames(schema.asyncFields(false)))

	schema.setOnlyFields("ID")
	assert.ElementsMatch(t, []string{"ID"}, filedNames(schema.availableFields()))

	schema.setOnlyFields("id")
	assert.ElementsMatch(t, []string{"ID"}, filedNames(schema.availableFields()))

	schema.setOnlyFields("ID", "NotFound")
	assert.ElementsMatch(t, []string{"ID"}, filedNames(schema.availableFields()))

	schema = newSchema(&UserSchema2{})
	schema.setExcludeFields("ID", "Name", "School")
	assert.ElementsMatch(t, []string{"Age", "Async"}, filedNames(schema.availableFields()))

	schema.setExcludeFields("ID", "Name", "School", "NotFound")
	assert.ElementsMatch(t, []string{"Age", "Async"}, filedNames(schema.availableFields()))
}

func filedNames(fields []*schemaField) (names []string) {
	for _, f := range fields {
		names = append(names, f.Name())
	}
	return
}
