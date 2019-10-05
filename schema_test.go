package portal

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type SchoolSchema struct {
	Name string
	Addr string
}

type PersonSchema struct {
	Age int
}

type UserSchema struct {
	PersonSchema
	ID     string
	Name   string        `portal:"meth:GetName"`
	School *SchoolSchema `portal:"nested"`
	Async  int           `portal:"async"`
}

func (u *UserSchema) GetName(user interface{}) interface{} {
	return "test"
}

func TestSchema(t *testing.T) {
	schema := NewSchema(&UserSchema{})
	assert.Equal(t, "UserSchema", schema.Name())
	assert.NotNil(t, schema.Struct())

	user := struct {
		ID     int
		School *SchoolSchema
	}{10, &SchoolSchema{Name: "test school"}}

	idField := NewField(schema, schema.schemaStruct.Field("ID"))
	val, err := schema.FieldValueFromData(context.TODO(), idField, user)
	assert.Nil(t, err)
	assert.Equal(t, 10, val)

	nameField := NewField(schema, schema.schemaStruct.Field("Name"))
	val, err = schema.FieldValueFromData(context.TODO(), nameField, user)
	assert.Nil(t, err)
	assert.Equal(t, "test", val)

	schoolField := NewField(schema, schema.schemaStruct.Field("School"))
	val, err = schema.FieldValueFromData(context.TODO(), schoolField, user)
	assert.Nil(t, err)
	assert.Equal(t, &SchoolSchema{Name: "test school", Addr: ""}, val)
}

func TestSchema_GetFields(t *testing.T) {
	schema := NewSchema(&UserSchema{})
	assert.ElementsMatch(t, []string{"Age", "ID", "Name", "School", "Async"}, filedNames(schema.AvailableFields()))

	assert.ElementsMatch(t, []string{"Age", "ID", "Name", "School"}, filedNames(schema.SyncFields()))
	assert.ElementsMatch(t, []string{"Async"}, filedNames(schema.AsyncFields()))

	schema.SetOnlyFields("ID")
	assert.ElementsMatch(t, []string{"ID"}, filedNames(schema.AvailableFields()))

	schema = NewSchema(&UserSchema{})
	schema.SetExcludeFields("ID", "Name", "School")
	assert.ElementsMatch(t, []string{"Age", "Async"}, filedNames(schema.AvailableFields()))
}

func filedNames(fields []*Field) (names []string) {
	for _, f := range fields {
		names = append(names, f.Name())
	}
	return
}
